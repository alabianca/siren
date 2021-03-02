# Siren
Transfer files on save.

## Example
Coding on your development machine, but transfering files to a raspberry pi.

## Usage
Listen on your raspberry pi (you can specifiy a port if you like). Files are transfered to the `pwd`.
```
    siren -listen
```

Now in your project directory run this command to listen of file save events and transfer them to your raspberry pi.
```
    siren -port <PORT_RPI_IS_LISTENING_ON> -host <IP_OF_RPI>
```