username=$1
password=$2
export TAG=$3
export IMAGE_URL='reminderimage'
export COMMIT_SHA=$(git rev-parse HEAD)


COLLECT_ERROR=True fandogh login --username $username --password $password

echo "image name: ${IMAGE_URL}"
echo "image version: ${TAG}"
echo "commit sha: ${TAG}"

COLLECT_ERROR=True fandogh image init --name  $IMAGE_URL
COLLECT_ERROR=True fandogh image publish --version $TAG

COLLECT_ERROR=True fandogh service apply -f ./deploy/fandogh/service.yml -p IMAGE_URL -p COMMIT_SHA -p TAG