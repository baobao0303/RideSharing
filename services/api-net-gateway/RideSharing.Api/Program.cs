using Yarp.ReverseProxy.Configuration;
using Microsoft.OpenApi.Models;
using RideSharing.Api.Services;
using RideSharing.Grpc.Auth;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.

// Add Swagger/OpenAPI
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen(c =>
{
    c.SwaggerDoc("v1", new OpenApiInfo
    {
        Title = "RideSharing API Gateway",
        Version = "v1",
        Description = "API Gateway for RideSharing Microservices Architecture. " +
                     "This gateway routes requests to the following services: " +
                     "Auth Service, Logger Service, and Mail Service.",
        Contact = new OpenApiContact
        {
            Name = "API Support",
            Email = "support@ridesharing.com"
        }
    });

    // Add security definition for JWT Bearer
    c.AddSecurityDefinition("Bearer", new OpenApiSecurityScheme
    {
        Description = "JWT Authorization header using the Bearer scheme. Example: \"Authorization: Bearer {token}\"",
        Name = "Authorization",
        In = ParameterLocation.Header,
        Type = SecuritySchemeType.ApiKey,
        Scheme = "Bearer"
    });

    c.AddSecurityRequirement(new OpenApiSecurityRequirement
    {
        {
            new OpenApiSecurityScheme
            {
                Reference = new OpenApiReference
                {
                    Type = ReferenceType.SecurityScheme,
                    Id = "Bearer"
                }
            },
            Array.Empty<string>()
        }
    });

    // Add API Gateway routes documentation
    c.TagActionsBy(api => new[] { api.GroupName ?? "API Gateway" });
});

// Add YARP Reverse Proxy
builder.Services.AddReverseProxy()
    .LoadFromConfig(builder.Configuration.GetSection("ReverseProxy"));

// gRPC clients
builder.Services.AddSingleton<RideSharing.Api.Services.GrpcClients>();

var app = builder.Build();

// Configure the HTTP request pipeline.
// Enable Swagger in all environments for API Gateway
app.UseSwagger();
app.UseSwaggerUI(c =>
{
        c.SwaggerEndpoint("/swagger/v1/swagger.json", "RideSharing API Gateway v1");
        c.RoutePrefix = "swagger"; // Swagger UI at /swagger
        c.DocumentTitle = "RideSharing API Gateway - Swagger UI";
    c.DefaultModelsExpandDepth(-1);
    c.DocExpansion(Swashbuckle.AspNetCore.SwaggerUI.DocExpansion.List);
});

// Health check endpoint
app.MapGet("/health", () => Results.Ok(new { status = "healthy", service = "api-gateway" }))
    .WithTags("Health")
    .WithName("HealthCheck")
    .Produces<object>(StatusCodes.Status200OK);

// API Gateway Information endpoint
app.MapGet("/api/info", () => Results.Ok(new
{
    service = "RideSharing API Gateway",
    version = "1.0.0",
    routes = new
    {
        auth = new
        {
            user = "/api/v1/User/*",
            email = "/api/v1/email/*",
            cities = "/api/v1/cities/*",
            token = "/api/v1/XFWToken/*"
        },
        logger = new
        {
            logs = "/api/v1/Logger/*"
        },
        mail = new
        {
            send = "/api/v1/Mail/*"
        }
    },
    swagger = "/swagger"
}))
    .WithTags("Info")
    .WithName("GatewayInfo")
    .Produces<object>(StatusCodes.Status200OK);

// HTTP -> gRPC bridge endpoints
app.MapGroup("/api/v1/Auth")
    .WithTags("Auth (HTTP->gRPC)")
    .MapAuthEndpoints();

// HTTP -> gRPC bridge endpoints for other services
app.MapGroup("/api/v1/Logger").WithTags("Logger (HTTP->gRPC)").MapLoggerEndpoints();
app.MapGroup("/api/v1/Mail").WithTags("Mail (HTTP->gRPC)").MapMailEndpoints();
app.MapGroup("/api/v1/Image").WithTags("Image (HTTP->gRPC)").MapImageEndpoints();

// Use YARP Reverse Proxy (kept for backward-compat)
app.MapReverseProxy();

app.Run();
