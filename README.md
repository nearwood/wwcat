Never tell me the odds.

Attempts to crack Warpwallet:
https://keybase.io/warp/warp_1.0.9_SHA256_a2067491ab582bde779f4505055807c2479354633a2216b22cf1e92d1a6e4a87.html

Uses my forked version of Go warpwallet implementation: https://github.com/nearwood/warpwallet

Can guess around 1 million guesses/day
130 days left till deadline
(130 million / 62^8) * 100 = 0.00006% chance

But hey it's fun. Also my first Go application.

I'm willing to bet the target password has at least one of each character class (upper, lower, digit),
so that reduces the search space a little:

62^8 - 36^8 - 36^8 - 52^8 + 10^8 + 26^8+ 26^8 - 26^8 - 26^8 - 10^8  
=159238157238528  
vs  
62^8  
=218340105584896  

Reduced by about 27%  

So now a 0.0000816% chance!

6f54dae - Initial version with history check  
d329573 - Fix DB connection for remote hosts  
33b1caf - Add kp/s (keypair/second) stat output every 5 keypairs  
e7a36bd - Inserts winning keypair if found with special date  
f5a1675 - Decided to add private key to DB for posterity  
b30286c - Decided to add public key and computer hostname while I was at it  
7584620 - Output total collisions incurred per instance, for curiosity's sake  
8b2e3d7 - No more collision check -- too slow.
84b50   - Truncate .git for Github
