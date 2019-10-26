#!/bin/bash
#
# Quick start script to get docker and docker-compose installed
# on an AWS instance using the default image.

# install docker and docker-compose
sudo yum update -y
sudo yum install -y python3-pip
sudo amazon-linux-extras install -y docker
sudo pip3 install docker-compose

# enable docker service and add user to the docker group
sudo usermod -a -G docker ec2-user
sudo systemctl enable docker
sudo systemctl start docker

# should log out and back for group to take effect
echo "log out and log back in!"
