@echo off

REM ===============================
REM Config - PROJECT_ID, REGION, REPO
REM ===============================
set PROJECT_ID=janna-tp
set REGION=asia-southeast1
set REPO=cloud-run-source-deploy
set IMAGE_NAME=tp-api-dev
set IMAGE=asia-southeast1-docker.pkg.dev/%PROJECT_ID%/%REPO%/%IMAGE_NAME%:latest
set SERVICE_NAME=%IMAGE_NAME%

cmd /c "gcloud config set project %PROJECT_ID%"


REM ===============================
REM Step 1: Build Docker image
REM ===============================
echo Building Docker image...
docker build -t %IMAGE% .

REM ===============================
REM Step 2: Push Docker image to Artifact Registry
REM ===============================
echo Pushing Docker image...
docker push %IMAGE%

REM ===============================
REM Step 3: Deploy to Cloud Run
REM ===============================
echo Deploying to Cloud Run...
gcloud run deploy %SERVICE_NAME% ^
  --image %IMAGE% ^
  --region %REGION% ^
  --platform managed ^
  --allow-unauthenticated ^
  --port 8080 ^
  --memory 512Mi ^
  --concurrency 80 ^

REM ===============================
echo Deployment finished!
echo Cloud Run URL:
gcloud run services describe %SERVICE_NAME% --region %REGION% --format "value(status.url)"
pause
