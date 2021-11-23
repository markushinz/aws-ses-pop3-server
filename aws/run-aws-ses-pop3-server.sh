PORT=$1

if [ -z $PORT ] ; then
  PORT=110
fi

export POP3_AUTHORIZATION_LAMBDA=EmailServerAuthenticationLambdaFunction
export POP3_AWS_ACCESS_KEY_ID=AKIARFD2USYQL5TA6RHJ
export POP3_AWS_SECRET_ACCESS_KEY=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
export POP3_AWS_S3_REGION=us-east-1
export POP3_AWS_S3_BUCKET=insite-development-device-email-data
export POP3_AWS_S3_PREFIX=Inbox/development.upload.powerside.com/
export POP3_PORT=${PORT}

cd /home/ec2-user/aws-ses-pop3-server

/usr/local/bin/go run main.go

