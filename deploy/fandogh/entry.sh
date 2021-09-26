username=$1
password=$2

export IMAGE_URL='reminderimage'
echo $RELEASE_VERSION
export TAG=$RELEASE_VERSION

COLLECT_ERROR=True fandogh login --username $username --password $password

COLLECT_ERROR=True fandogh image init --name  $IMAGE_URL
COLLECT_ERROR=True fandogh image publish --version $RELEASE_VERSION

COLLECT_ERROR=True fandogh service apply -f ./deploy/fandogh/service.yml -p IMAGE_URL -p TAG