# Intro 

`onlinestat` is a online user counter service.

Each time user interact with website, log ip with a exipred time in redis. Then, the count of keys is the result we need.

NOTE: no auth check provided, so DO NOT expose your service to public.

# LICENSE

MIT