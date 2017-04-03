This program is intended to be run on two devices which are communicating over 6LoWPAN. One device should run the server program, 
while the other should run the client program. The server broadcasts messages to the clients, and the client receives the message. 
The messages are all encrypted using an AES 256 algorithm. The encryption key is a public key that is generated with the Diffie-Hellman
method. The messages are simply reandom strings that are generated to test the code. Several messages may be decryped or encryped at a 
time, ensuring high throughput. The encryption key changes each second to ensure high security.
