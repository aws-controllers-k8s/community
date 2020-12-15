FROM python:3.8-slim AS integ-runner

WORKDIR /root/tests

# Install python dependencies
COPY ./requirements.txt .
RUN pip install -r requirements.txt

# Set up default environment variables
ENV PYTHONPATH=/root/tests

COPY . .
RUN mkdir -p $HOME/.kube

# Run the tests
ENTRYPOINT ["./run-tests.sh"]