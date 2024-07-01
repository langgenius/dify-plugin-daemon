FROM ubuntu:22.04

# Install conda
RUN apt-get update && apt-get install -y wget && \
    wget https://repo.anaconda.com/miniconda/Miniconda3-py39_4.10.3-Linux-x86_64.sh && \
    bash Miniconda3-py39_4.10.3-Linux-x86_64.sh -b -p /miniconda3 && \
    rm Miniconda3-py39_4.10.3-Linux-x86_64.sh