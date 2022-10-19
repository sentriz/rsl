<h3 align=center><b>rsl</b></h3>
<p align=center><i><b>r</b>e<b>s</b>eria<b>l</b>ise: lossy but versatile conversion between data serialisation formats </i></p>

---

### installation

```shell
    $ go install go.senan.xyz/rsl@latest
```

### usage

##### command line

```shell
    $ rsl <src format> <dest format>
```

##### available formats

- **csv**
- **csv-ph** (csv with generated pseudo-header)
- **js** (javascript objects, decode only)
- **json**
- **toml**
- **xml** (lossy support for arbitrary objects)
- **xml-std** (lossless but limited)
- **yaml**

### examples

```shell
    $ rsl toml csv <some-toml.toml >some-csv.csv
```

```shell
    $ cat example.json
    [
        {
          "name": "jim",
          "addr": "dublin"
        },
        {
          "name": "miguel",
          "addr": "space"
        }
    ]
```

```shell
    $ rsl json json <example.json
    [{"addr":"dublin","name":"jim"},{"addr":"space","name":"miguel"}]
```

```shell
    $ rsl json toml <example.json
    [[result]]
      addr = "dublin"
      name = "jim"

    [[result]]
      addr = "space"
      name = "miguel"
```

```shell
    $ rsl json yaml <example.json
    - addr: dublin
      name: jim
    - addr: space
      name: miguel
```

```shell
    $ rsl json csv <example.json
    name,addr
    jim,dublin
    miguel,space
```

```shell
    $ cat example.simple.json
    [
        [
          "jim",
          "dublin"
        ],
        [
          "miguel",
          "space"
        ]
    ]
```

```shell
    # generate pseudo headers if there are none. makes it easy to query with other tools
    $ rsl json csv <example.simple.json
    a,b
    jim,dublin
    miguel,space
```

```shell
    $ cat example.html
    <select id="cars">
      <option value="volvo">Volvo</option>
      <option value="saab">Saab</option>
      <option value="opel">Opel</option>
      <option value="audi">Audi</option>
    </select>
```

```shell
    $ rsl xml json <example.html | jq
    {
      "select": {
        "@id": "cars",
        "option": [
          {
            "#text": "Volvo",
            "@value": "volvo"
          },
          {
            "#text": "Saab",
            "@value": "saab"
          },
          {
            "#text": "Opel",
            "@value": "opel"
          },
          {
            "#text": "Audi",
            "@value": "audi"
          }
        ]
      }
    }
```

```shell
    $ rsl xml json <example.html | jq -r '.select.option | .[] | select(."@value" == "saab") | ."#text"'
    Saab
```

```shell
    $ cat example.csv
    year,artist,album
    1981,Alan Vega,"Collision Drive"
    1991,Underground Resistance,"The Final Frontier"
    1999,Dopplereffekt,Gesamtkunstwerk
    2002,Dopplereffekt,Myon-Neutrino
    2016,Pangaea,"In Drum Play"
```

```shell
    $ rsl csv json <example.csv | jq -r '.[] | select(.artist == "Dopplereffekt") | .album'
    Gesamtkunstwerk
    Myon-Neutrino
```

```shell
    # parse javascript objects
    $ node -e 'console.log({colour: "blue", v: Math.random()})' | rsl js yaml
    colour: blue
    v: 0.7409782831156317
```
