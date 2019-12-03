## download data
tarPath="./google-data"
path="./google-data"
sudo mkdir $path
sudo chmod 777 $path
sudo mkdir $tarPath
sudo chmod 777 $tarPath

wget --load-cookies /tmp/cookies.txt \
  "https://docs.google.com/uc?export=download&confirm=$(wget --quiet --save-cookies /tmp/cookies.txt --keep-session-cookies --no-check-certificate 'https://docs.google.com/uc?export=download&id=1HhrFN03QfZQ9HzI0Y9OXuMZXliUvN2cp' -O- | sed -rn 's/.*confirm=([0-9A-Za-z_]+).*/\1\n/p')&id=1HhrFN03QfZQ9HzI0Y9OXuMZXliUvN2cp" \
  -O $tarPath/tasks-res.tar && rm -rf /tmp/cookies.txt
sudo tar -xvf $tarPath/tasks-res.tar -C $path 
rm -rf $tarPath/tasks-res.tar
mv $path/tasks-res $path/tasks

wget --load-cookies /tmp/cookies.txt \
  "https://docs.google.com/uc?export=download&confirm=$(wget --quiet --save-cookies /tmp/cookies.txt --keep-session-cookies --no-check-certificate 'https://docs.google.com/uc?export=download&id=1ymFRBvW1wKIHrdi-v5wyzJLKu81ZxFOx' -O- | sed -rn 's/.*confirm=([0-9A-Za-z_]+).*/\1\n/p')&id=1ymFRBvW1wKIHrdi-v5wyzJLKu81ZxFOx" \
  -O $tarPath/machines.tar && rm -rf /tmp/cookies.txt
sudo tar -xvf $tarPath/machines.tar -C $path
rm -rf $tarPath/machines.tar