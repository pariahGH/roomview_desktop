# Roomview_Desktop

Desktop version of the Roomview_Mobile repo, actually works because I don't have to worry about Android doing funky things with my background processes.

Expects rooms.csv to be a CSV file with room name as first column and IPv4 address as second column as well as any file named alert.wav to be next the executable otherwise it will just ignore you when you run it.

Alert gets looped until you hit the clear alerts button. 

As I poke at Roomview more I may add features to this just like with the mobile version, including the ability to remotely shut off systems - a power that will of course be used responsibly.

Uses [malgo](https://github.com/youpy/go-wav) for audio playback and [golang-ui/nuklear](https://github.com/golang-ui/nuklear) for the GUI