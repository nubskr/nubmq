Goal: Outperform rabbitmq in terms of pure throughput without sacrificing reliability in any way 

Issues(non fatal ones):

- Switches are being triggered a lot more frequently than they need to for some reason
- Some keys and values are somehow being inserted in both tables, something really weird is happening there for some reason(migration might be the reason btw,maybe we are somehow inserting keys in new table and also migrating them somehow from old to new)

TODO(fatal stuff):

- Data races happen at times, we need more locks at places dumbo