## download data
tarPath="./google-data"; \
path="./google-data"; \
sudo mkdir $path; \
sudo chmod 777 $path; \
sudo mkdir $tarPath; \
sudo chmod 777 $tarPath

wget --load-cookies /tmp/cookies.txt \
  "https://docs.google.com/uc?export=download&confirm=$(wget --quiet --save-cookies /tmp/cookies.txt --keep-session-cookies --no-check-certificate 'https://docs.google.com/uc?export=download&id=1HhrFN03QfZQ9HzI0Y9OXuMZXliUvN2cp' -O- | sed -rn 's/.*confirm=([0-9A-Za-z_]+).*/\1\n/p')&id=1HhrFN03QfZQ9HzI0Y9OXuMZXliUvN2cp" \
  -O $tarPath/tasks-res.tar && rm -rf /tmp/cookies.txt; \
sudo tar -xvf $tarPath/tasks-res.tar -C $path; \
rm -rf $tarPath/tasks-res.tar; \
mv $path/tasks-res $path/tasks

# https://drive.google.com/file/d/1tvEBcB9gJMtMV5T2jwRxm2cO9Um59Dxd/view?usp=sharing
wget --load-cookies /tmp/cookies.txt \
  "https://docs.google.com/uc?export=download&confirm=$(wget --quiet --save-cookies /tmp/cookies.txt --keep-session-cookies --no-check-certificate 'https://docs.google.com/uc?export=download&id=1tvEBcB9gJMtMV5T2jwRxm2cO9Um59Dxd' -O- | sed -rn 's/.*confirm=([0-9A-Za-z_]+).*/\1\n/p')&id=1tvEBcB9gJMtMV5T2jwRxm2cO9Um59Dxd" \
  -O $tarPath/tasks-res-mani.tar && rm -rf /tmp/cookies.txt; \
sudo tar -xvf $tarPath/tasks-res-mani.tar -C $path; \
rm -rf $tarPath/tasks-res-mani.tar; \
mv $path/tasks-res $path/tasks

wget --load-cookies /tmp/cookies.txt \
  "https://docs.google.com/uc?export=download&confirm=$(wget --quiet --save-cookies /tmp/cookies.txt --keep-session-cookies --no-check-certificate 'https://docs.google.com/uc?export=download&id=1ymFRBvW1wKIHrdi-v5wyzJLKu81ZxFOx' -O- | sed -rn 's/.*confirm=([0-9A-Za-z_]+).*/\1\n/p')&id=1ymFRBvW1wKIHrdi-v5wyzJLKu81ZxFOx" \
  -O $tarPath/machines.tar && rm -rf /tmp/cookies.txt; \
sudo tar -xvf $tarPath/machines.tar -C $path; \
rm -rf $tarPath/machines.tar 

## google trace


## install gsutil
echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list; \
sudo apt-get install apt-transport-https ca-certificates gnupg; \
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add - ;\
sudo apt-get update && sudo apt-get install google-cloud-sdk; \
sudo apt-get install google-cloud-sdk-app-engine-java; \
gcloud init


## download
gsutil cp -R gs://clusterdata-2011-2 ./
