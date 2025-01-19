Goal: Be a beast outperforming everything that comes in its way

# TODO:

## Core engine changes:
- Optimize for speed

## New feats:
- pub sub notifs


Problem:

we are relying on usedCapacity of SM to trigger upgrades and downgrades, but it is not reliable at all

make the totalCapacity and usedCapacity a *int from int

### commands supported rn:
```
SET key value
SET key value EX ex_time_in_seconds

GET key
```