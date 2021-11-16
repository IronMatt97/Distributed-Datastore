# README / HOWTO INSTALL
Per avviare il codice sarà sufficiente utilizzare degli script bash appositamente realizzati.
All'interno della directory Install&Run sarà possibile trovare un file chiamato install.sh, avviando il quale, verrà invocato docker per la creazione delle immagini dei container necessari a far girare l'applicativo. Avviando invece run.sh specificando il numero di datastore e di api, verranno avviati i container necessari al sistema.
Ovviamente il requisito necessario all'avvio del sistema è Docker, senza il quale non potranno essere create le immagini.
##Guida:
###1. Spostarsi nella cartella Install&Run (cd Install&Run)
###2. Eseguire lo script di installazione (./install.sh)
###3. Eseguire lo script di avvio (./run.sh [numeroDatastore] [numeroAPI])
