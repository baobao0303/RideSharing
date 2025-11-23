using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using RideSharing.Grpc.Logger;

namespace RideSharing.Api.Services;

public static class LoggerEndpointsExtensions
{
    public static RouteGroupBuilder MapLoggerEndpoints(this RouteGroupBuilder group)
    {
        // POST /api/v1/Logger/log
        group.MapPost("/log", async (GrpcClients clients, LogRequest req) =>
        {
            var reply = await clients.LoggerClient.LogInfoAsync(req);
            return Results.Ok(reply);
        })
        .WithName("Logger_LogInfo");

        return group;
    }
}

