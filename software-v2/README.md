# software v2

the original software was getting unweildy. this is a full reimplementation in C, using micropython as the scripting lanaguee

to build: run `./build.sh`. It autoflashes with picotool. make sure everything is installed.

this new version is made for v2. there are a few minor references to v1 in the code so parts of it may work, but these are for my own testing only and the system is designed with v2 in mind

# todo
[X] integration with micropython and game loop system with update and init functions
[] file system
[] sprite drawing
[] gpio
[] dirty drawing


# dirty rectangles
first pass - does it already fit in one of the existing rectangles?
secnd pass - what is the most optimal dirty rect to add it to at this point
    if this fits above a certain threshold of how many pixels it would save, we take that and extend that
    if not, we create a new dirty rectangle
    if we're already above the quota of dirty rectangles, we go back to the one which would be the most optimal (even if above threshold) and add it to that
    
if we make a rectangle larger - chain reaction, see if some other 

# Monitoring serial
the thing shows up as a serial port that you can connect to. it gives you 3 seconds to connect on boot, and then runs the program

can use normal terminal utils for this. i use tio in a secondary split screen terminal, which auto reconnects when the device comes back. on mac:

`tio /dev/cu.usbmodem*`

you will need to install tio with either your linux PM or homebrew