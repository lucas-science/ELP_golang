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
