# What is this?
A plan on how to code snake. I want to try and figure out the logic without any copilot or anything.
i want this to be more efficient than it needs to be. 

main struct
- current direction
- slice of {x,y} coordinates
  - can't use normal slices since we'll be constantly adding to the end and taking away from the start, and exhaust ram by just moving the pointer forwards a ton
  - apparently a ring buffer is a thing. I think i'll choose this - it comes out to 1.1kb of ram for the whole thing
  - linked list


params
- board width: 24*24, cells 10 pixels each


color 
- changes colour depending on which level it is in

rgb565 colours - format of RRRRRGGGGGGBBBBB where each one is a number
