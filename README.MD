# USCIS CHECK STATUS

This tool allows you to check your case status directly from the terminal. Alternatively, you can deploy it as an amazon lambda and have the updates to be sent to your phone.

## But wait, why don't you just check on the actual page

Hey, why can't we apply a bit of overengineering to daily tasks?!
Also, I wanted to play with two things that I applied here (the Go port of nokogiri and lambdas), so I had the perfect excuse out of a common lambda tutorial :P

# Usage

- Download the binary from the releases page
- Add execution permissions `chmod +x ./uscis`
- Execute like below
```shell
./uscis MSC999999999
```

# Collaboration

Fork, code, push, PR :)

# License

## License

Copyright © 2019 [Thiago Costa](mailto:thiago@arroadie.com)
This work is free. You can redistribute it and/or modify it under the
terms of the Do What The Fuck You Want To Public License, Version 2,
as published by Sam Hocevar. See the COPYING file for more details.
