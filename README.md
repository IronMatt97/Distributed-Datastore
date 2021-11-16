# README / HOWTO INSTALL
Per avviare il codice sarà sufficiente utilizzare degli script bash appositamente realizzati.
All'interno della directory Install&Run sarà possibile trovare un file chiamato install.sh, avviando il quale, verrà invocato docker per la creazione delle immagini dei container necessari a far girare l'applicativo. Avviando invece run.sh specificando il numero di datastore e di api, verranno avviati i container necessari al sistema.
Ovviamente il requisito necessario all'avvio del sistema è Docker, senza il quale non potranno essere create le immagini.

Guida per l'avvio della demo:
1. Spostarsi nella cartella Install&Run (cd Install&Run)
2. Eseguire lo script di installazione (./install.sh)
3. Eseguire lo script di avvio (./run.sh [numeroDatastore] [numeroAPI])
4. NOTA: Qualora non si avesse a disposizione il terminale chiamato Konsole (di Kubuntu), lo script di avvio non funzionerà e basterà sostituire nello script la stringa "konsole" con la stringa del proprio terminale preferito. In alternativa basta avviare manualmente i container, con i comandi docker run -it [NomeImmagine] (i nomi immagine preinstallati dal primo script saranno client, api, datastore e discovery, verificabile con il comando docker image ls)

A questo punto si potrà utilizzare il nodo client per effettuare delle richieste al sistema, seguendo la guida che comparirà a schermo.

Note aggiuntive:
Qualora si voglia testare il sistema in maniera approfondita è stato lasciato un Dockerfile nella certella NodeInspector_Dockerfile.
Utilizzando quest'ultimo è possibile creare un'immagine in grado di simulare qualsiasi nodo con una connessione ssh attiva instaurata, così da avere maggior libertà di gestione.
Guida alla modalità di ispezione:
1. Spostarsi nella cartella NodeInspector_Dockerfile (cd NodeInspector_Dockerfile)
2. Richiedere a Docker di creare l'immagine (docker build --tag node .)
3. Avviare un numero di container a piacere (docker run -it node)

A questo punto sarà possibile scegliere che funzionalità il nodo dovrà ricoprire entrando nella cartella SDCC del container, cosi da tenere sotto controllo l'output del nodo, con la possibilità di fermare e farne ripartire il funzionamento, o di modificarne il codice. A tal proposito sono già installati nel container strumenti utili.
