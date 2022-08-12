# go-updater
Yet another go updater/installer. Main purpose is just upgrade current golang installed version.

## Usage
Just build it and execute it or run it as a script.
```Shell
go build -o go-updater && sudo ./go-updater
```
or
```Shell
sudo go run .
```

## Considerations
Super user permissions are needed in order to write at /usr/local (where the installation files live).

It is also worth mentioning that this is a Lignux only project. The script won't work on other operating systems as it was not written with those in mind. I am sorry if that is inconvenient but I don't think it will.

You won't be able to build or run the project if you have not any go version installed obviously. In that case, please, build it anywhere first, then execute it on the desired system.
