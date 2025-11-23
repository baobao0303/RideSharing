//go:build grpc
// +build grpc

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	authpb "ride-sharing/shared/generated/auth"
	"ride-sharing/services/auth/data"
)

const grpcPort = 50000

// authServer implements the AuthService gRPC server
// It uses the same data.Models as the HTTP handlers
// to keep a single source of truth for business logic.

type authServer struct {
	authpb.UnimplementedAuthServiceServer
	models *data.Models
}

func startGRPCServer(models *data.Models) {
	addr := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}

	s := grpc.NewServer()
	authpb.RegisterAuthServiceServer(s, &authServer{models: models})
	// Enable server reflection for easier debugging
	reflection.Register(s)

	log.Printf("gRPC AuthService listening on %s", addr)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()
}

// ===== Helper: JWT generation & validation =====
func jwtSecret() string {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		sec = "your-secret-key-change-in-production"
	}
	return sec
}

func generateAccessToken(userID, email, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
		"iat":   time.Now().Unix(),
		"type":  "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func generateRefreshToken(userID, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
		"type": "refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func parseAndValidateJWT(tokenString string, expectedType string) (*jwt.Token, jwt.MapClaims, error) {
	tok, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret()), nil
	})
	if err != nil {
		return nil, nil, err
	}

	if !tok.Valid {
		return tok, nil, fmt.Errorf("token invalid")
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return tok, nil, fmt.Errorf("invalid claims")
	}
	if expectedType != "" {
		if ctype, _ := claims["type"].(string); ctype != expectedType {
			return tok, claims, fmt.Errorf("unexpected token type: %s", ctype)
		}
	}
	return tok, claims, nil
}

// ===== gRPC Methods =====

func (s *authServer) SignUp(ctx context.Context, req *authpb.SignUpRequest) (*authpb.SignUpResponse, error) {
	if s.models == nil {
		return nil, fmt.Errorf("models not initialized")
	}

	if req.GetFirstName() == "" || req.GetLastName() == "" || req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, fmt.Errorf("firstName, lastName, email, and password are required")
	}
	if len(req.GetPassword()) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters long")
	}

	// Parse optional cityId & genderId
	var cityID *int
	if req.GetCityId() != "" {
		if id, err := strconv.Atoi(req.GetCityId()); err == nil {
			// verify city
			if _, err := s.models.City.GetOne(id); err == nil {
				cityID = &id
			}
		}
	}
	var genderID *int
	if req.GetGenderId() != "" {
		if id, err := strconv.Atoi(req.GetGenderId()); err == nil {
			if _, err := s.models.Gender.GetByID(id); err == nil {
				genderID = &id
			}
		}
	}

	// Parse DOB optional
	var dob *time.Time
	if req.GetDob() != "" {
		if d, err := time.Parse(time.RFC3339, req.GetDob()); err == nil {
			dob = &d
		} else if d2, err2 := time.Parse("2006-01-02", req.GetDob()); err2 == nil {
			dob = &d2
		} else {
			return nil, fmt.Errorf("invalid DOB format")
		}
	}

	// Ensure not exists
	if _, err := s.models.User.GetByEmail(req.GetEmail()); err == nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	u := data.User{
		Email:           req.GetEmail(),
		FirstName:       req.GetFirstName(),
		MiddleName:      nilIfEmpty(req.GetMiddleName()),
		LastName:        req.GetLastName(),
		Password:        req.GetPassword(),
		Active:          true,
		CityID:          cityID,
		GenderID:        genderID,
		DOB:             dob,
		InvitationToken: nilIfEmpty(req.GetInvitationToken()),
	}
	id, err := s.models.User.Insert(u)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &authpb.SignUpResponse{
		UserId:    id,
		Email:     req.GetEmail(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		Message:   "User registered successfully",
	}, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *authServer) SignIn(ctx context.Context, req *authpb.SignInRequest) (*authpb.SignInResponse, error) {
	if s.models == nil {
		return nil, fmt.Errorf("models not initialized")
	}
	if req.GetUserName() == "" || req.GetPassword() == "" {
		return nil, fmt.Errorf("userName and password are required")
	}

	user, err := s.models.User.GetByEmail(req.GetUserName())
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !user.Active {
		return nil, fmt.Errorf("user account is inactive")
	}

	tmp := data.User{Password: user.Password}
	ok, err := tmp.PasswordMatches(req.GetPassword())
	if err != nil || !ok {
		return nil, fmt.Errorf("invalid credentials")
	}

	access, err := generateAccessToken(user.ID, user.Email, jwtSecret())
	if err != nil {
		return nil, fmt.Errorf("error generating access token")
	}
	refresh, err := generateRefreshToken(user.ID, jwtSecret())
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token")
	}

	return &authpb.SignInResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    3600,
		UserId:       user.ID,
		UserName:     user.Email,
	}, nil
}

func (s *authServer) VerifyAccessToken(ctx context.Context, req *authpb.VerifyAccessTokenRequest) (*authpb.VerifyAccessTokenResponse, error) {
	if req.GetToken() == "" {
		return &authpb.VerifyAccessTokenResponse{Valid: false, Message: "Token is required"}, nil
	}
	_, _, err := parseAndValidateJWT(req.GetToken(), "access")
	if err != nil {
		return &authpb.VerifyAccessTokenResponse{Valid: false, Message: err.Error()}, nil
	}
	return &authpb.VerifyAccessTokenResponse{Valid: true, Message: "Token is valid"}, nil
}

func (s *authServer) RenewAccessToken(ctx context.Context, req *authpb.RenewAccessTokenRequest) (*authpb.RenewAccessTokenResponse, error) {
	if req.GetVrto() == "" {
		return nil, fmt.Errorf("vrto (refresh token) is required")
	}
	_, claims, err := parseAndValidateJWT(req.GetVrto(), "refresh")
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %v", err)
	}
	uid, _ := claims["sub"].(string)
	email, _ := claims["email"].(string) // may be empty
	access, err := generateAccessToken(uid, email, jwtSecret())
	if err != nil {
		return nil, fmt.Errorf("error generating access token")
	}
	return &authpb.RenewAccessTokenResponse{AccessToken: access, RefreshToken: req.GetVrto(), ExpiresIn: 3600}, nil
}

func (s *authServer) VerifyMail(ctx context.Context, req *authpb.VerifyMailRequest) (*authpb.VerifyMailResponse, error) {
	if s.models == nil {
		return nil, fmt.Errorf("models not initialized")
	}
	if req.GetVerificationOTPCode() == "" || req.GetOtp() == "" {
		return nil, fmt.Errorf("verificationOTPCode and otp are required")
	}
	vt, err := s.models.VerifyToken.GetByToken(req.GetVerificationOTPCode())
	if err != nil {
		return nil, fmt.Errorf("invalid verification token")
	}
	if vt.IsUsed {
		return nil, fmt.Errorf("verification token already used")
	}
	if time.Now().After(vt.ExpiresAt) {
		return nil, fmt.Errorf("verification token expired")
	}
	if vt.OTPCode != req.GetOtp() {
		return nil, fmt.Errorf("invalid OTP code")
	}
	if err := s.models.VerifyToken.MarkAsUsed(vt.ID); err != nil {
		return nil, fmt.Errorf("error updating verification status")
	}
	return &authpb.VerifyMailResponse{Message: "Email verified successfully", UserId: vt.UserID}, nil
}

func (s *authServer) ResendOTP(ctx context.Context, req *authpb.ResendOTPRequest) (*authpb.ResendOTPResponse, error) {
	if s.models == nil {
		return nil, fmt.Errorf("models not initialized")
	}
	if req.GetVerificationOTPCode() == "" {
		return nil, fmt.Errorf("verificationOTPCode is required")
	}
	vt, err := s.models.VerifyToken.GetByToken(req.GetVerificationOTPCode())
	if err != nil {
		return nil, fmt.Errorf("verification token not found")
	}
	if vt.IsUsed {
		return nil, fmt.Errorf("verification token already used")
	}
	if time.Now().After(vt.ExpiresAt) {
		return nil, fmt.Errorf("verification token expired")
	}
	newToken := fmt.Sprintf("%d", time.Now().UnixNano())
	newOTP := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	_, err = s.models.VerifyToken.Insert(data.VerifyToken{
		UserID:           vt.UserID,
		Token:            newToken,
		OTPCode:          newOTP,
		VerificationType: vt.VerificationType,
		ExpiresAt:        time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return nil, fmt.Errorf("error generating new verification token")
	}
	return &authpb.ResendOTPResponse{Message: "OTP sent successfully", VerificationToken: newToken}, nil
}

func (s *authServer) GetProvinces(ctx context.Context, req *authpb.GetProvincesRequest) (*authpb.GetProvincesResponse, error) {
	list, err := s.models.City.GetAllProvinces()
	if err != nil {
		return nil, err
	}
	res := &authpb.GetProvincesResponse{}
	for _, c := range list {
		res.Provinces = append(res.Provinces, &authpb.City{
			Id:           int32(c.ID),
			Code:         c.Code,
			Name:         c.Name,
			Type:         c.Type,
			ProvinceCode: safeStr(c.ProvinceCode),
			ParentCode:   safeStr(c.ParentCode),
		})
	}
	return res, nil
}

func (s *authServer) GetWards(ctx context.Context, req *authpb.GetWardsRequest) (*authpb.GetWardsResponse, error) {
	list, err := s.models.City.GetWardsByProvinceCode(req.GetProvinceCode())
	if err != nil {
		return nil, err
	}
	res := &authpb.GetWardsResponse{}
	for _, c := range list {
		res.Wards = append(res.Wards, &authpb.City{
			Id:           int32(c.ID),
			Code:         c.Code,
			Name:         c.Name,
			Type:         c.Type,
			ProvinceCode: safeStr(c.ProvinceCode),
			ParentCode:   safeStr(c.ParentCode),
		})
	}
	return res, nil
}

func safeStr(s string) string { return s }

