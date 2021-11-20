using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net;
using System.Threading.Tasks;
using Azure.Storage.Blobs;
using Azure.Storage.Queues;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using Microsoft.Extensions.Logging;

namespace PublicApi.Func
{
    public static class ArrProxy
    {
        [Function("ArrProxy")]
        public static async Task<HttpResponseData> Run(
            [HttpTrigger(AuthorizationLevel.Function, "get", "post")] HttpRequestData req,
            FunctionContext executionContext)
        {
            var logger = executionContext.GetLogger("ArrProxy");
            logger.LogInformation("C# HTTP trigger function processed a request.");

            var body = await req.ReadAsStringAsync();

            await StoreRequest(body);
            await QueueRequest(body);
            
            var response = req.CreateResponse(HttpStatusCode.OK);
            response.Headers.Add("Content-Type", "text/plain; charset=utf-8");
            await response.WriteStringAsync("Processed request.");
            return response;
        }

        private static async Task StoreRequest(string request)
        {
            var bsc = new BlobServiceClient(Environment.GetEnvironmentVariable("STORAGE_ACCOUNT"));
            var container = bsc.GetBlobContainerClient("webhook-requests");
            var blob = container.GetBlobClient($"{DateTime.Now:yyyy-MM-dd hh:mm:ss.fff}.json");

            await using var stream = new MemoryStream();
            var writer = new StreamWriter(stream);
            await writer.WriteAsync(request);
            await writer.FlushAsync();
            stream.Position = 0;
            
            await blob.UploadAsync(stream);
        }

        private static async Task QueueRequest(string request)
        {
            var queueClient = new QueueClient(Environment.GetEnvironmentVariable("STORAGE_ACCOUNT"), "webhook-requests");
            
            // Wait for autoscan wait time and plex scan to complete before making message available to consumers
            await queueClient.SendMessageAsync(request, visibilityTimeout: TimeSpan.FromMinutes(15));
        }
    }
}