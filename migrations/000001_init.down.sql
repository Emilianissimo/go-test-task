upstream app_upstream {
    server app:8080 fail_timeout=0;
}

server {
    listen 8080;
    server_name localhost;

    location / {
        proxy_pass http://app_upstream;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}