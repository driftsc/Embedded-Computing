Lab 0 Readme.

This lab covers the dining philosopher's problem as described in the Lab 0 assignment.
the program prints the philosopher's name and activity if the philosopher changes activities.
(i.e. if the philosopher was thinking and starts thinking about something else, then nothing
is printed. If he was eating and is now thinking, then it prints out his name and that he is
now thinking.)
In order to run this program, follow these steps:

Open Lab0.exe
The program will start and execute.
To exit, press Enter, or close the window.

The source code is available in lab0.go which has internal notes to describe how the
program was made.

The program uses four Goroutines as philosophers, and the forks (chopsticks) are placed
between them. The forks are shared resources and are handled as channels between the
four philosophers.

The order around the table in a clockwise manner is as follows:
philosopher0
fork0
philosopher1
fork1
philosopher2
fork2
philosopher3
fork3
(philosopher0)

Each philosopher is instructed to pick up the lower number fork first. This means that
philosophers 1, 2, and 3 will pick up the fork to their right, then the one to their left,
while philosopher 0 will pick up the fork to his left, then the one to his right.

This scheme prevents a deadlock because two philosophers try to pick up fork 0 first,
and they cannot both have fork0, so there must always be a free fork which a different
philosopher can use to eat. Once that philosopher is finished eating, he will free up
one of the remaining forks for the remaining philosophers, so that each one will have a turn to eat.
The philosophers cannot all have a fork at the same time, so there is no deadlock condition.

if a fork is in use, the philosopher will wait a half of a second before checking on the
fork again, preventing a busy-wait issue.
