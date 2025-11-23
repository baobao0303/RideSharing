using Grpc.Net.Client;
using RideSharing.Grpc.Auth;
using RideSharing.Grpc.Logger;
using RideSharing.Grpc.Mail;
using RideSharing.Grpc.Image;

namespace RideSharing.Api.Services;

public class GrpcClients
{
    public AuthService.AuthServiceClient AuthClient { get; }
    public LoggerService.LoggerServiceClient LoggerClient { get; }
    public MailService.MailServiceClient MailClient { get; }
    public ImageService.ImageServiceClient ImageClient { get; }

    public GrpcClients(IConfiguration configuration)
    {
        // Allow HTTP/2 without TLS for gRPC (h2c)
        AppContext.SetSwitch("System.Net.Http.SocketsHttpHandler.Http2UnencryptedSupport", true);

        var httpHandler = new SocketsHttpHandler { EnableMultipleHttp2Connections = true };

        var authUrl = configuration["Grpc:AuthService:Url"] ?? "http://auth:50000";
        var loggerUrl = configuration["Grpc:LoggerService:Url"] ?? "http://logger-service:50001";
        var mailUrl = configuration["Grpc:MailService:Url"] ?? "http://mail-service:50002";
        var imageUrl = configuration["Grpc:ImageService:Url"] ?? "http://image-service:50003";

        var authCh = GrpcChannel.ForAddress(authUrl, new GrpcChannelOptions { HttpHandler = httpHandler });
        var loggerCh = GrpcChannel.ForAddress(loggerUrl, new GrpcChannelOptions { HttpHandler = httpHandler });
        var mailCh = GrpcChannel.ForAddress(mailUrl, new GrpcChannelOptions { HttpHandler = httpHandler });
        var imageCh = GrpcChannel.ForAddress(imageUrl, new GrpcChannelOptions { HttpHandler = httpHandler });

        AuthClient = new AuthService.AuthServiceClient(authCh);
        LoggerClient = new Logger.LoggerService.LoggerServiceClient(loggerCh);
        MailClient = new Mail.MailService.MailServiceClient(mailCh);
        ImageClient = new Image.ImageService.ImageServiceClient(imageCh);
    }
}
