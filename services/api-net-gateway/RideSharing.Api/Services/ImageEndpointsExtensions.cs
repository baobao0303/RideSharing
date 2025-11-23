using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using RideSharing.Grpc.Image;

namespace RideSharing.Api.Services;

public static class ImageEndpointsExtensions
{
    public static RouteGroupBuilder MapImageEndpoints(this RouteGroupBuilder group)
    {
        // POST /api/v1/Image/upload/folder (multipart form-data)
        // form fields: folder, file (IFormFile), optional fileName
        group.MapPost("/upload/folder", async (GrpcClients clients, HttpRequest httpRequest) =>
        {
            if (!httpRequest.HasFormContentType)
                return Results.BadRequest(new { error = "Content-Type must be multipart/form-data" });

            var form = await httpRequest.ReadFormAsync();
            var folder = form["folder"].ToString();
            var file = form.Files["file"];
            var fileName = form["fileName"].ToString();

            if (string.IsNullOrWhiteSpace(folder) || file is null)
                return Results.BadRequest(new { error = "folder and file are required" });

            if (string.IsNullOrWhiteSpace(fileName)) fileName = file.FileName;

            await using var ms = new MemoryStream();
            await file.CopyToAsync(ms);

            var req = new UploadRequest
            {
                Folder = folder,
                FileName = fileName,
                Content = Google.Protobuf.ByteString.CopyFrom(ms.ToArray()),
                ContentType = file.ContentType ?? "application/octet-stream"
            };

            var reply = await clients.ImageClient.UploadToFolderAsync(req);
            return Results.Ok(reply);
        })
        .DisableAntiforgery()
        .WithName("Image_UploadToFolder");

        return group;
    }
}

