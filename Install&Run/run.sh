#!/bin/bash
if [ -z "$1" ] || [ -z "$2" ] || ! [[ "$1" =~ ^[0-9]+$ ]] || ! [[ "$2" =~ ^[0-9]+$ ]] ||[ "$#" -ne 2 ]
then
    	echo "Errore: Gli argomenti inseriti non sono validi. Eseguire lo script come ./run #Datastore #Restapi"
	exit -1
fi
printf -v int '%d\n' "$1" 2>/dev/null
printf -v int '%d\n' "$2" 2>/dev/null

echo "-------------------------------------------"
echo "AVVIO DEL SERVIZIO DI DATASTORE DISTRIBUITO"
echo "-------------------------------------------"
echo "# DISCOVERY NODE	= 1"
echo "# DATASTORE NODES	= $1"
echo "# RESTAPI NODES   = $2"
sleep 1

echo "Avvio del nodo di discovery ..."
konsole --noclose -e docker run -it discovery & disown
sleep 3

for ((i = 1; i<=$1; i++));
do
	echo "Avvio del Datastore richiesto..."
	konsole --noclose -e docker run -it datastore & disown
	sleep 1
done

for ((i = 1; i<=$2; i++));
do
	echo "Avvio della API richiesta..."
	konsole --noclose -e docker run -it api & disown
	sleep 1
done

echo "Avvio del nodo client ..."
konsole --noclose -e docker run -it client & disown
echo "SISTEMA AVVIATO CORRETTAMENTE."
exit 0
