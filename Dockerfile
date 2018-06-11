FROM busybox

COPY ./mikrodns /home/

ENTRYPOINT [ "/home/mikrodns" ] 