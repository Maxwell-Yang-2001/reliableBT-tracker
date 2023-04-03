# You can run this script to restart the tracker service with the latest changes and config.

./trakx stop
# force the new yaml to be used
rm ~/.config/trakx/trakx.yaml
# build with latest changes
go build
./trakx start