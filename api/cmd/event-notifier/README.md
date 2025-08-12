# Event Notifier for IBM&reg; Power&reg; Access Cloud

## Overview
The Event Notifier is a system that sends notifications for various events on a IBM&reg; Power&reg; Access Cloud portal, such as group membership requests, approvals, and rejections etc. 

## Integration with Notification Service
Notification integration is pretty flexible and can be implemented in any way. Right now, we have implemented email notifications via the [Email Delivery][sendgrid] service in the IBM Cloud.

## DB requirements

Event notifier is utilizing the mongo db [watch][watch-mongo] collection feature for watching the new entry and acts on it. In order work this feature mongodb collection needs to be capped and following instruction is for setting the cap on the events collection.

```shell
# connect to the mongosh
$ mongosh "mongodb://username:password@host:27017/"
test> use pac
pac> db.createCollection("events", { capped: true, size: 300000 })
{ ok: 1 }
pac>
```


[sendgrid]: https://cloud.ibm.com/catalog/infrastructure/email-delivery
[watch-mongo]: https://www.mongodb.com/docs/manual/changeStreams/#watch-a-collection--database--or-deployment
