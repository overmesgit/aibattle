docker build -t mycompiler ./
docker tag mycompiler asia-northeast1-docker.pkg.dev/cookies-444312/cloud-run-source-deploy/mycompiler
docker push asia-northeast1-docker.pkg.dev/cookies-444312/cloud-run-source-deploy/mycompiler
gcloud run deploy mycompiler --region asia-northeast1 \
--image asia-northeast1-docker.pkg.dev/cookies-444312/cloud-run-source-deploy/mycompiler \
--port 8080 \
--memory 2Gi \
--cpu 4 \
--env-vars-file env.yaml
