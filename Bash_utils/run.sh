konsole --noclose -e docker run -it discovery & disown
sleep 3
konsole --noclose -e docker run -it datastore & disown
konsole --noclose -e docker run -it datastore & disown
konsole --noclose -e docker run -it api & disown
konsole --noclose -e docker run -it api & disown
konsole --noclose -e docker run -it client & disown
