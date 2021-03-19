# graffitiwallpainter

`graffitiwallpainter` will periodically fetch the current [graffitiwall](https://beaconcha.in/graffitiwall) 
from beaconcha.in and check which pixels need to be painted to draw the given image on the graffitiwall 
and updates a [graffiti-file](https://docs.prylabs.network/docs/prysm-usage/graffiti-file/).


```
beaconcha.in/graffitiwall -> graffitiwallpainter -> graffiti.yml -> validator
                             ^
                             |
/path/to/image.png ----------+
```

## usage

```bash
# install the binary
go install github.com/guybrush/graffitiwallpainter@latest
# offset x:10 y:10 only fetch once
graffitiwallpainter -image /path/to/image.png -graffiti /path/to/graffiti-file.yml -x 10 -y 10 -once
# draw on pyrmont-graffitiwall
graffitiwallpainter -url https://pyrmont.beaconcha.in/graffitiwall -image /path/to/image.png -graffiti /path/to/graffiti-file.yml -x 10 -y 10 -once

# run with docker (it will update the file every minute)
docker run -v $PWD:/v guybrush/graffitiwallpainter -image /v/image.png -graffiti /v/graffiti.yml -x 100 -y 100
```
