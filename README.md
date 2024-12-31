# Standup Notes

A tool to parse markdown notes and generate:

* Yesterday's completed items/worked on
* Today's planned items

Prints out in slack-compatible markdown so you can copy/paste into slack

# TODO

* [x] remove breaks from standup notes + update template
* [ ] update links to match the actual previous rather than blind yesterday
* [ ] Fix line-wrapping missing spaces in the output
* [ ] Allow for top level journal/standup dirs to be variably named

## Conventions

* Journal/daily notes are named in `YYYY-DD-MM.md` format
* Standup notes are named in `YYYY-DD-MM.md` format
* Standup/Journal notes are stored in sibling directories (e.g `notes/standup` and `notes/journal`)
* Standup and journal top level directories are hardcoded to be named `standup` and `journal` respectively

## Workflow

Daily goals and work done recorded in a journal note (along with other information)

1. At the beginning of the day generate a new journal note
    * Update the links to previous days journal
    * Copy over the goals of the day from the previous day and add/modify/remove as needed
    * Copy over the goals of the week from the previous day
1. Generate a standup note
    * Update the links to previous days journal and standup
    * Extract work done from the previous days journal to the work done section
    * Extract work planned for the day from the current day's journal to the today section
1. Export the standup note into slack

