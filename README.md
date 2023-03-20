# Intro 

`onlinestat` is a online user counter service.

Currently support two methods

- Redis

Log ip with a exipred time in redis DB. Then, the count of keys is the result we need.

- Memory

Log ip using golang map. In this mode, you SHOULD only deploy one instance.

NOTE: no auth check provided, so DO NOT expose your service to public.

# LICENSE

MIT