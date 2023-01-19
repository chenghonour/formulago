FROM golang:1.19-alpine

LABEL maintainer="coko@duck.com"

###############################################################################
#                                INSTALLATION
###############################################################################
# Set project path
ENV WORKDIR /var/www/formulago
# Add the application executable and set the execution permission
ADD ./formulago   $WORKDIR/formulago
RUN chmod +x $WORKDIR/formulago

###############################################################################
#                                   START
###############################################################################
WORKDIR $WORKDIR
# Set the environment variables
ENV IS_PROD true
# Run the application
CMD ./formulago