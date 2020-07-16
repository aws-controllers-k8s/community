# Frequently Asked Questions (FAQ)

## Service Broker

!!! question "Question"
    Does ACK replace the [service broker](https://svc-cat.io/)?

!!! quote "Answer"
    For the time being, people using the service broker should continue to use it and we're coordinating with the maintainers to provide a unified solution.

    The service broker project is also an AWS activity that, with the general shift of focus in the community from service broker to operators, can be considered less actively developed. There are a certain things around application lifecycle management that the service broker currently covers and which are at this juncture not yet covered by the scope of ACK, however we expect in the mid to long run that these two projects converge. We had AWS-internal discussions with the team that maintains the service broker and we're on the same page concerning a unified solution.

    We appreciate input and advice concerning features that are currently covered by the service broker only, for example bind/unbind or cataloging and looking forward to learn from the community how they are using service broker so that we can take this into account.

## Contributing

!!! question "Question"
    Where and how can I help?

!!! quote "Answer"
    Excellent question and we're super excited that you're interested in ACK.
    For now, if you're a developer, you can check out the [mvp](https://github.com/aws/aws-controllers-k8s/tree/mvp) branch and try out the code generation.