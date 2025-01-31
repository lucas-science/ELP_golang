# ELP/Golang : Filtrage de similarité dans un ensemble de photo

## Ce qu'on a fait

- Interface graphique
- Envois photo client => sevreur
- Réception photo par le serveur
- Algo testant les similarité entre toutes les images
- Envois photo serveur => client
- Réception photo par le client

## Ce qu'il nous reste à faire

- fix le bug aleatoire : desfois des segments sont pas de la bonne taille attendue donc ça plante (=> mettre en place une logique de renvois si erreur sur un segment)
- fix l'algo pour que la similarité soit plus flagrante (on a une distance de pixel encore trop proche entre les cas : deux photo différentes et deux photo similaire)

Commandes unix à installer :
  sudo apt install build-essential libx11-dev libgl1-mesa-dev xorg-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev
  
  Aller dans le repertoire où on veut installer OpenCV et GoCV
  faire : git clone https://github.com/hybridgroup/gocv.git   puis faire
	sudo apt update
	sudo apt install -y build-essential cmake pkg-config
	sudo apt install -y libopencv-dev
	sudo apt install -y libjpeg-dev libpng-dev libtiff-dev
	sudo apt install -y libgtk-3-dev
  puis ensuite "cd gocv" et "make install"
  
  
