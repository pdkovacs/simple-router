Frontend to a multi-instance service which can be used to route requests
based on values in a configurable header or cookie. For example, making sure
that requests of different users will always be serviced by differrent instances
helps test use cases where the assumption/precondition is that user are not
guaranteed to share the same instance. An example of such a use case is when a
user's action trigger real-time notifications to be sent to other users.
