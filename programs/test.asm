// Load 10 into R0
LOD R0 10

// Loop until R0 is 0
#loop
SUB R0 1
BRZ R0 #end
BRN #loop

#end
BRN #end