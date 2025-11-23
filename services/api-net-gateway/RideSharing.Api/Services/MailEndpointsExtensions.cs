using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using RideSharing.Grpc.Mail;

namespace RideSharing.Api.Services;

public static class MailEndpointsExtensions
{
    public static RouteGroupBuilder MapMailEndpoints(this RouteGroupBuilder group)
    {
        // POST /api/v1/Mail/send
        group.MapPost("/send", async (GrpcClients clients, MailRequest req) =>
        {
            var reply = await clients.MailClient.SendMailAsync(req);
            return Results.Ok(reply);
        })
        .WithName("Mail_Send");

        return group;
    }
}
