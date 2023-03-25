FROM oven/bun:latest

ENV PORT 3000
EXPOSE 3000

ADD . /home/bun/app
CMD ["/usr/local/bin/bun", "run", "index.ts"]