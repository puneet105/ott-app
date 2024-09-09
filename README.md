## Project Overview

In this project, we aimed to build a video streaming service with support for multiple resolutions and formats, including HLS and DASH protocols. Here’s a summary of the tasks and how they were implemented:

### Initial Setup and Manifest Generation

- Implemented a Go server to handle video streaming using `ffmpeg` to generate video manifests.
- Created endpoints to start the manifest generation and stream the generated video content.

### Asynchronous Manifest Generation

- The `/generate` endpoint starts manifest generation in a Goroutine. This allows the server to respond immediately with a task ID while the manifest generation runs in the background.
- The (`/status/{taskID}`) to check the progress of the manifest generation. The status could be "started", "completed", or "failed".

### Handling Multiple Resolutions and Formats

- Enhanced the manifest generation to support different video resolutions and formats (HLS and DASH). Each resolution and format was handled in separate Goroutines for improved performance.
- Used `ffmpeg` commands configured for both HLS and DASH formats, leveraging hardware acceleration features on Apple Silicon (e.g., `h264_videotoolbox` codec).

### Path and CORS Configuration

- Adjusted server and client configurations to ensure that paths to the video files and manifests are correctly handled and accessible.
- Configured CORS headers to support cross-origin requests, enabling smoother integration with web clients.

# Testing the Video Streaming Application

## Prerequisites

- Ensure that `ffmpeg` is installed and properly configured on your system.
- Ensure that your Go environment is set up and the application dependencies are installed.

## Testing the Application

### 1. Run the Application

```
go run main.go

```

#### Obtain a JWT Token
You need to log in to obtain a JWT token that will be used for subsequent authenticated requests.
1. Login to Get a JWT Token:Use curl to make a POST request to the /login endpoint.

```
curl -X POST -d ""username":"puneet", "password":"puneet123"" http://localhost:8080/login
```

2. If the credentials are correct, you'll receive a JWT token as a response.

```
 User exists, Generating Token ...!!!
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InB1bmVldCIsImV4cCI6MTcyNTI5ODM4NX0.o4qkjDN8CaA-AbApHlfJSZEL6qZvmyxdt8F6YC0oT88
```

3. Save the JWT Token:Save this token for use in the subsequent steps.

#### To Generate the manigest Jobs to the Queues (Redis and RabbitMQ)
Use the JWT token obtained from the login to generate the mani.
1. Publish a Job:Use curl to make a POST request to the /generate endpoint.

```
curl -X POST -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNzIzODk4NjM0fQ.V7C8reg3zfYH14rSF5FnT70jox-Lb-N4XMT2B6LxTsw" http://localhost:8080/generate
```

2. Replace <your_jwt_token> with the token you obtained earlier.

3. Check the Response:If successful, you should see a below response

```
status: Manifest generation started
 taskID: 436dc6f7-22ea-4a66-ba50-0967553511f1
```

#### Check Manifest Generation Status
Once you have the task ID from the previous step, check the status of the manifest generation by sending a GET request to the /status/{taskID} endpoint:

```
curl -X GET http://localhost:8080/status/436dc6f7-22ea-4a66-ba50-0967553511f1

```
Expected Response: The server should return the status of the manifest generation.

```
{
    "taskID": "436dc6f7-22ea-4a66-ba50-0967553511f1",
    "status": "InProgress...!!!"
}
```

#### Stream Video Content
To stream the video content, ensure that the manifests are generated and available in the manifests directory. Access the video stream by sending a GET request to the /stream endpoint:

For HLS:

```
http://localhost:8080/stream/hls/360p/manifest
http://localhost:8080/stream/hls/480p/manifest
http://localhost:8080/stream/hls/720p/manifest
http://localhost:8080/stream/hls/1080p/manifest
```

For DASH:

```
http://localhost:8080/stream/dash/360p/manifest
http://localhost:8080/stream/dash/480p/manifest
http://localhost:8080/stream/dash/720p/manifest
http://localhost:8080/stream/dash/1080p/manifest
```

# How to play .m3u8 and .mpd files locally.
To test whether above generated .m3u8 and .mpd files are being created properly, Open vlc playes on MacBook and follow below options to play video using generated .m3u8 and .mpd files.

**File > OpenFile  > SELECT .m3u8 and .mpd file**

Now you should be able to play video this way.

## Troubleshooting

1. 404 Not Found: Verify that the manifest and .ts files exist in the correct directory. Check the paths in the application configuration.
2. Manifest Generation Delays: Ensure that ffmpeg is properly optimized and that Goroutines are correctly handling parallel tasks.
3. CORS Issues: Check the server's CORS configuration and make sure that allowed origins and headers are correctly set.

