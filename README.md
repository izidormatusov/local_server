local_server
============

Avoid procrastination and speedup productive pages

## Motivation

Do you visit some websites more than you'd like? Are there some websites, on the other hand, that are really painful to get to? `local_server` saves the day!

## Installation

On Ubuntu:

  make install

and restart your computer.

## Default configuration

Shortcuts (defined in `REDIRECTS`):

 - [c/](http://c/) goes to Google Calendar
 - [m/](http://m/) goes to Google Inbox
 - [d/](http://d/) goes to Google Drive

Blocked websites (defined in `ANTI_PROCRASTINATION`):

 - 9gag.com
 - facebook.com
 - news.ycombinator.com
 - twitter.com

Instead of the blocked website you can see one of:

 - a disgusting [anti-soda ad](http://www.youtube.com/embed/Naj5NIVl4mw)
 - a picture of the [fat gamer guy](https://www.youtube.com/watch?v=MT0-OL_71yU) from South Park
 - a cute picture of a kitty
 - a cute picture of a husky
 - a picture from reddit:
   - [/r/Pictures](http://reddit.com/r/Pictures)
   - [/r/funny](http://reddit.com/r/funny)
   - [/r/EarthPorn](http://reddit.com/r/EarthPorn)
   - [/r/CityPorn](http://reddit.com/r/CityPorn)
   - [/r/AnimalPorn](http://reddit.com/r/AnimalPorn)
   - [/r/itookapicture](http://reddit.com/r/itookapicture)

## How does it work?

`local_server -install` installs aliases into `/etc/hosts` pointing to address `127.0.42.42`. When you run `local_server`, it listens on `127.0.42.42:80` and serves replacement content.

## Credits

[Izidor Matušov](http://izidor.io) <[izidor.matusov@gmail.com](mailto:izidor.matusov@gmail.com)>
