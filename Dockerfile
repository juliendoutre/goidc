FROM scratch

LABEL org.opencontainers.image.authors Julien Doutre <jul.doutre@gmail.com>
LABEL org.opencontainers.image.title goidc
LABEL org.opencontainers.image.url https://github.com/juliendoutre/goidc
LABEL org.opencontainers.image.documentation https://github.com/juliendoutre/goidc
LABEL org.opencontainers.image.source https://github.com/juliendoutre/goidc
LABEL org.opencontainers.image.licenses MIT

COPY goidc /
ENTRYPOINT ["/goidc"]
