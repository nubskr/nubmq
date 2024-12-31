Goal: Outperform rabbitmq in terms of pure throughput without sacrificing reliability in any way 

Issues(non fatal ones):

- Locks are not optimized enough, things can be tweaked a lot more, they just *work* now, they can be so much faster and better

TODO(fatal stuff):

- Some get requests are failing, it seems like they keys are just not there in the bucket where they are supposed to be, seems like some gap in our set system which is letting some keys fall through due to concurrency
