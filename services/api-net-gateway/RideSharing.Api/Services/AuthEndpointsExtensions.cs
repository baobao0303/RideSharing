using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using RideSharing.Grpc.Auth;

namespace RideSharing.Api.Services;

public static class AuthEndpointsExtensions
{
    public static RouteGroupBuilder MapAuthEndpoints(this RouteGroupBuilder group)
    {
        // POST /api/v1/Auth/sign-up
        group.MapPost("/sign-up", async (GrpcClients clients, SignUpRequest req) =>
        {
            var reply = await clients.AuthClient.SignUpAsync(req);
            return Results.Ok(reply);
        })
        .WithName("Auth_SignUp");

        // POST /api/v1/Auth/sign-in
        group.MapPost("/sign-in", async (GrpcClients clients, SignInRequest req) =>
        {
            var reply = await clients.AuthClient.SignInAsync(req);
            return Results.Ok(reply);
        })
        .WithName("Auth_SignIn");

        // POST /api/v1/Auth/verify-mail
        group.MapPost("/verify-mail", async (GrpcClients clients, VerifyMailRequest req) =>
        {
            var reply = await clients.AuthClient.VerifyMailAsync(req);
            return Results.Ok(reply);
        })
        .WithName("Auth_VerifyMail");

        // POST /api/v1/Auth/resend-otp
        group.MapPost("/resend-otp", async (GrpcClients clients, ResendOTPRequest req) =>
        {
            var reply = await clients.AuthClient.ResendOTPAsync(req);
            return Results.Ok(reply);
        })
        .WithName("Auth_ResendOTP");

        // GET /api/v1/Auth/verify-access-token?token=...
        group.MapGet("/verify-access-token", async (GrpcClients clients, string? token) =>
        {
            var reply = await clients.AuthClient.VerifyAccessTokenAsync(new VerifyAccessTokenRequest { Token = token ?? string.Empty });
            return Results.Ok(reply);
        })
        .WithName("Auth_VerifyAccessToken");

        // POST /api/v1/Auth/renew-access-token
        group.MapPost("/renew-access-token", async (GrpcClients clients, RenewAccessTokenRequest req) =>
        {
            var reply = await clients.AuthClient.RenewAccessTokenAsync(req);
            return Results.Ok(reply);
        })
        .WithName("Auth_RenewAccessToken");

        // Cities
        group.MapGet("/cities/provinces", async (GrpcClients clients) =>
        {
            var reply = await clients.AuthClient.GetProvincesAsync(new GetProvincesRequest());
            return Results.Ok(reply);
        })
        .WithName("Auth_GetProvinces");

        group.MapGet("/cities/provinces/{province_code}/wards", async (GrpcClients clients, string province_code) =>
        {
            var reply = await clients.AuthClient.GetWardsAsync(new GetWardsRequest { ProvinceCode = province_code });
            return Results.Ok(reply);
        })
        .WithName("Auth_GetWards");

        return group;
    }
}
