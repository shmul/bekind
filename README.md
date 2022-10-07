# bekind

## What is it
Just a toy project to add some useful online utilities supported by DNS and HTTPS requests.

## _Features_

### caller ip address
try `{ip,myip,my}.zif.im` e.g.
```
dig @34.148.107.114  +short ip.zif.im
"46.117.241.248"
```

### random unique id
try `{id,key}.zif.im` (based on [nanoid](https://github.com/jaevor/go-nanoid)), e.g.
```
dig @34.148.107.114  +short key.zif.im
"2y29jkma6bfw"
```

Default length is 12, but `N.id.zif.im` will return `N` lengthed random string (when I implement it).


### URL Shortner
Pending
