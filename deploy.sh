KEY_PATH="$HOME/.ssh/id_ed25519"
REMOTE_HOST="root@138.197.182.90"
REMOTE_APP_DIR="/opt/touchly"
BUILD_FOLDER="build"

swag init -g cmd/api/main.go

cd cmd/api && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o main
cd ../..
mkdir -p $BUILD_FOLDER
mv cmd/api/main "$BUILD_FOLDER/main"
cd $BUILD_FOLDER && zip -r main.zip main
cd ../

scp -i "$KEY_PATH" $BUILD_FOLDER/main.zip $REMOTE_HOST:$REMOTE_APP_DIR/main.zip
scp -i "$KEY_PATH" config.production.yaml $REMOTE_HOST:$REMOTE_APP_DIR/config.yaml
scp -i "$KEY_PATH" -r scripts/migrations $REMOTE_HOST:$REMOTE_APP_DIR
scp -i "$KEY_PATH" -r templates $REMOTE_HOST:$REMOTE_APP_DIR

ssh -i "$KEY_PATH" $REMOTE_HOST 'sudo systemctl stop touchly.service'
ssh -i "$KEY_PATH" $REMOTE_HOST '/opt/touchly/migrate -path /opt/touchly/migrations -database postgres://postgres:mysecretpassword@localhost:5432/touchly?sslmode=disable up'
ssh -i "$KEY_PATH" $REMOTE_HOST 'cd /opt/touchly; unzip -o main.zip; sudo systemctl restart touchly.service'
rm -rf $BUILD_FOLDER
