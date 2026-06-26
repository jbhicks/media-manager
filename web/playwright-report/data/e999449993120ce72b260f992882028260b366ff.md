# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: media-manager.spec.ts >> Media Manager - Discover Page >> should load discover page with content
- Location: e2e/media-manager.spec.ts:56:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('text=Trending Movies')
Expected: visible
Error: strict mode violation: locator('text=Trending Movies') resolved to 2 elements:
    1) <p class="text-[#b3b3b3] text-lg">Explore trending movies and TV shows</p> aka getByText('Explore trending movies and')
    2) <h2 class="text-xl font-bold text-white">Trending Movies</h2> aka getByRole('heading', { name: 'Trending Movies' })

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('text=Trending Movies')

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - generic [ref=e5]:
      - button [ref=e6] [cursor=pointer]:
        - img [ref=e7]
      - generic [ref=e9]: 📺 Media Manager
      - generic [ref=e10]:
        - generic "VPN is not active" [ref=e11]:
          - img [ref=e12]
          - generic [ref=e14]: VPN Disconnected
        - button [ref=e15] [cursor=pointer]:
          - img [ref=e16]
        - button [ref=e22] [cursor=pointer]:
          - img [ref=e23]
  - generic [ref=e26]:
    - navigation [ref=e27]:
      - generic [ref=e28]:
        - link "Home" [ref=e29] [cursor=pointer]:
          - /url: /
          - img [ref=e30]
          - text: Home
        - link "Discover" [ref=e33] [cursor=pointer]:
          - /url: /discover
          - img [ref=e34]
          - text: Discover
        - link "Watchlist" [ref=e37] [cursor=pointer]:
          - /url: /watchlist
          - img [ref=e38]
          - text: Watchlist
        - link "Downloads" [ref=e40] [cursor=pointer]:
          - /url: /downloads
          - img [ref=e41]
          - text: Downloads
        - link "Library" [ref=e44] [cursor=pointer]:
          - /url: /library
          - img [ref=e45]
          - text: Library
        - link "Search" [ref=e47] [cursor=pointer]:
          - /url: /search
          - img [ref=e48]
          - text: Search
        - link "Suggestions" [ref=e51] [cursor=pointer]:
          - /url: /suggestions
          - img [ref=e52]
          - text: Suggestions
        - link "Settings" [ref=e54] [cursor=pointer]:
          - /url: /settings
          - img [ref=e55]
          - text: Settings
    - main [ref=e58]:
      - generic [ref=e59]:
        - img [ref=e60]
        - generic [ref=e62]:
          - paragraph [ref=e63]: VPN Disconnected
          - paragraph [ref=e64]: Downloads are disabled for your security. Please connect to your VPN to resume downloading.
        - button "Dismiss warning" [ref=e65] [cursor=pointer]:
          - img [ref=e66]
      - generic [ref=e69]:
        - generic [ref=e73]:
          - heading "Discover" [level=1] [ref=e74]
          - paragraph [ref=e75]: Explore trending movies and TV shows
        - generic [ref=e77]:
          - generic [ref=e78]:
            - button "All" [ref=e79] [cursor=pointer]
            - button "Movies" [ref=e80] [cursor=pointer]
            - button "TV Shows" [ref=e81] [cursor=pointer]
          - button "Filters" [ref=e83] [cursor=pointer]:
            - img [ref=e84]
            - text: Filters
        - generic [ref=e86]:
          - generic [ref=e87]:
            - generic [ref=e88]:
              - img [ref=e90]
              - heading "Trending Movies" [level=2] [ref=e93]
              - img [ref=e94]
            - generic [ref=e96]:
              - link "Toy Story 5 7.4 Toy Story 5 2026" [ref=e97] [cursor=pointer]:
                - /url: /movie/1084244
                - generic [ref=e98]:
                  - img "Toy Story 5" [ref=e99]
                  - generic [ref=e102]:
                    - img [ref=e103]
                    - generic [ref=e105]: "7.4"
                - heading "Toy Story 5" [level=3] [ref=e106]
                - paragraph [ref=e107]: "2026"
              - link "The Sheep Detectives 7.6 The Sheep Detectives 2026" [ref=e108] [cursor=pointer]:
                - /url: /movie/1301421
                - generic [ref=e109]:
                  - img "The Sheep Detectives" [ref=e110]
                  - generic [ref=e113]:
                    - img [ref=e114]
                    - generic [ref=e116]: "7.6"
                - heading "The Sheep Detectives" [level=3] [ref=e117]
                - paragraph [ref=e118]: "2026"
              - link "Obsession 7.9 Obsession 2026" [ref=e119] [cursor=pointer]:
                - /url: /movie/1339713
                - generic [ref=e120]:
                  - img "Obsession" [ref=e121]
                  - generic [ref=e124]:
                    - img [ref=e125]
                    - generic [ref=e127]: "7.9"
                - heading "Obsession" [level=3] [ref=e128]
                - paragraph [ref=e129]: "2026"
              - link "Backrooms 6.9 Backrooms 2026" [ref=e130] [cursor=pointer]:
                - /url: /movie/1083381
                - generic [ref=e131]:
                  - img "Backrooms" [ref=e132]
                  - generic [ref=e135]:
                    - img [ref=e136]
                    - generic [ref=e138]: "6.9"
                - heading "Backrooms" [level=3] [ref=e139]
                - paragraph [ref=e140]: "2026"
              - link "Masters of the Universe 7.2 Masters of the Universe 2026" [ref=e141] [cursor=pointer]:
                - /url: /movie/454639
                - generic [ref=e142]:
                  - img "Masters of the Universe" [ref=e143]
                  - generic [ref=e146]:
                    - img [ref=e147]
                    - generic [ref=e149]: "7.2"
                - heading "Masters of the Universe" [level=3] [ref=e150]
                - paragraph [ref=e151]: "2026"
              - link "Michael 8.7 Michael 2026" [ref=e152] [cursor=pointer]:
                - /url: /movie/936075
                - generic [ref=e153]:
                  - img "Michael" [ref=e154]
                  - generic [ref=e157]:
                    - img [ref=e158]
                    - generic [ref=e160]: "8.7"
                - heading "Michael" [level=3] [ref=e161]
                - paragraph [ref=e162]: "2026"
              - link "Voicemails for Isabelle 8.2 Voicemails for Isabelle 2026" [ref=e163] [cursor=pointer]:
                - /url: /movie/614945
                - generic [ref=e164]:
                  - img "Voicemails for Isabelle" [ref=e165]
                  - generic [ref=e168]:
                    - img [ref=e169]
                    - generic [ref=e171]: "8.2"
                - heading "Voicemails for Isabelle" [level=3] [ref=e172]
                - paragraph [ref=e173]: "2026"
              - link "Disclosure Day 6.7 Disclosure Day 2026" [ref=e174] [cursor=pointer]:
                - /url: /movie/1275779
                - generic [ref=e175]:
                  - img "Disclosure Day" [ref=e176]
                  - generic [ref=e179]:
                    - img [ref=e180]
                    - generic [ref=e182]: "6.7"
                - heading "Disclosure Day" [level=3] [ref=e183]
                - paragraph [ref=e184]: "2026"
              - link "Supergirl 6.2 Supergirl 2026" [ref=e185] [cursor=pointer]:
                - /url: /movie/1081003
                - generic [ref=e186]:
                  - img "Supergirl" [ref=e187]
                  - generic [ref=e190]:
                    - img [ref=e191]
                    - generic [ref=e193]: "6.2"
                - heading "Supergirl" [level=3] [ref=e194]
                - paragraph [ref=e195]: "2026"
              - link "Scary Movie 5.5 Scary Movie 2026" [ref=e196] [cursor=pointer]:
                - /url: /movie/1273221
                - generic [ref=e197]:
                  - img "Scary Movie" [ref=e198]
                  - generic [ref=e201]:
                    - img [ref=e202]
                    - generic [ref=e204]: "5.5"
                - heading "Scary Movie" [level=3] [ref=e205]
                - paragraph [ref=e206]: "2026"
          - generic [ref=e207]:
            - generic [ref=e208]:
              - img [ref=e210]
              - heading "Popular Movies" [level=2] [ref=e212]
              - img [ref=e213]
            - generic [ref=e215]:
              - link "Cocktail 2 5.6 Cocktail 2 2026" [ref=e216] [cursor=pointer]:
                - /url: /movie/1392469
                - generic [ref=e217]:
                  - img "Cocktail 2" [ref=e218]
                  - generic [ref=e221]:
                    - img [ref=e222]
                    - generic [ref=e224]: "5.6"
                - heading "Cocktail 2" [level=3] [ref=e225]
                - paragraph [ref=e226]: "2026"
              - link "Obsession 7.9 Obsession 2026" [ref=e227] [cursor=pointer]:
                - /url: /movie/1339713
                - generic [ref=e228]:
                  - img "Obsession" [ref=e229]
                  - generic [ref=e232]:
                    - img [ref=e233]
                    - generic [ref=e235]: "7.9"
                - heading "Obsession" [level=3] [ref=e236]
                - paragraph [ref=e237]: "2026"
              - link "Madness 5.1 Madness 1980" [ref=e238] [cursor=pointer]:
                - /url: /movie/28322
                - generic [ref=e239]:
                  - img "Madness" [ref=e240]
                  - generic [ref=e243]:
                    - img [ref=e244]
                    - generic [ref=e246]: "5.1"
                - heading "Madness" [level=3] [ref=e247]
                - paragraph [ref=e248]: "1980"
              - link "Michael 8.7 Michael 2026" [ref=e249] [cursor=pointer]:
                - /url: /movie/936075
                - generic [ref=e250]:
                  - img "Michael" [ref=e251]
                  - generic [ref=e254]:
                    - img [ref=e255]
                    - generic [ref=e257]: "8.7"
                - heading "Michael" [level=3] [ref=e258]
                - paragraph [ref=e259]: "2026"
              - link "Mortal Kombat II 8.0 Mortal Kombat II 2026" [ref=e260] [cursor=pointer]:
                - /url: /movie/931285
                - generic [ref=e261]:
                  - img "Mortal Kombat II" [ref=e262]
                  - generic [ref=e265]:
                    - img [ref=e266]
                    - generic [ref=e268]: "8.0"
                - heading "Mortal Kombat II" [level=3] [ref=e269]
                - paragraph [ref=e270]: "2026"
              - link "Toy Story 5 7.4 Toy Story 5 2026" [ref=e271] [cursor=pointer]:
                - /url: /movie/1084244
                - generic [ref=e272]:
                  - img "Toy Story 5" [ref=e273]
                  - generic [ref=e276]:
                    - img [ref=e277]
                    - generic [ref=e279]: "7.4"
                - heading "Toy Story 5" [level=3] [ref=e280]
                - paragraph [ref=e281]: "2026"
              - link "Bhooth Bangla 5.6 Bhooth Bangla 2026" [ref=e282] [cursor=pointer]:
                - /url: /movie/1239134
                - generic [ref=e283]:
                  - img "Bhooth Bangla" [ref=e284]
                  - generic [ref=e287]:
                    - img [ref=e288]
                    - generic [ref=e290]: "5.6"
                - heading "Bhooth Bangla" [level=3] [ref=e291]
                - paragraph [ref=e292]: "2026"
              - link "Your Heart Will Be Broken 7.0 Your Heart Will Be Broken 2026" [ref=e293] [cursor=pointer]:
                - /url: /movie/1523145
                - generic [ref=e294]:
                  - img "Your Heart Will Be Broken" [ref=e295]
                  - generic [ref=e298]:
                    - img [ref=e299]
                    - generic [ref=e301]: "7.0"
                - heading "Your Heart Will Be Broken" [level=3] [ref=e302]
                - paragraph [ref=e303]: "2026"
              - link "Deep Water 6.1 Deep Water 2026" [ref=e304] [cursor=pointer]:
                - /url: /movie/1127384
                - generic [ref=e305]:
                  - img "Deep Water" [ref=e306]
                  - generic [ref=e309]:
                    - img [ref=e310]
                    - generic [ref=e312]: "6.1"
                - heading "Deep Water" [level=3] [ref=e313]
                - paragraph [ref=e314]: "2026"
              - link "Damage 6.6 Damage 1992" [ref=e315] [cursor=pointer]:
                - /url: /movie/11012
                - generic [ref=e316]:
                  - img "Damage" [ref=e317]
                  - generic [ref=e320]:
                    - img [ref=e321]
                    - generic [ref=e323]: "6.6"
                - heading "Damage" [level=3] [ref=e324]
                - paragraph [ref=e325]: "1992"
          - generic [ref=e326]:
            - generic [ref=e327]:
              - img [ref=e329]
              - heading "Now Playing" [level=2] [ref=e331]
              - img [ref=e332]
            - generic [ref=e334]:
              - link "Cocktail 2 5.6 Cocktail 2 2026" [ref=e335] [cursor=pointer]:
                - /url: /movie/1392469
                - generic [ref=e336]:
                  - img "Cocktail 2" [ref=e337]
                  - generic [ref=e340]:
                    - img [ref=e341]
                    - generic [ref=e343]: "5.6"
                - heading "Cocktail 2" [level=3] [ref=e344]
                - paragraph [ref=e345]: "2026"
              - link "Obsession 7.9 Obsession 2026" [ref=e346] [cursor=pointer]:
                - /url: /movie/1339713
                - generic [ref=e347]:
                  - img "Obsession" [ref=e348]
                  - generic [ref=e351]:
                    - img [ref=e352]
                    - generic [ref=e354]: "7.9"
                - heading "Obsession" [level=3] [ref=e355]
                - paragraph [ref=e356]: "2026"
              - link "Michael 8.7 Michael 2026" [ref=e357] [cursor=pointer]:
                - /url: /movie/936075
                - generic [ref=e358]:
                  - img "Michael" [ref=e359]
                  - generic [ref=e362]:
                    - img [ref=e363]
                    - generic [ref=e365]: "8.7"
                - heading "Michael" [level=3] [ref=e366]
                - paragraph [ref=e367]: "2026"
              - link "Mortal Kombat II 8.0 Mortal Kombat II 2026" [ref=e368] [cursor=pointer]:
                - /url: /movie/931285
                - generic [ref=e369]:
                  - img "Mortal Kombat II" [ref=e370]
                  - generic [ref=e373]:
                    - img [ref=e374]
                    - generic [ref=e376]: "8.0"
                - heading "Mortal Kombat II" [level=3] [ref=e377]
                - paragraph [ref=e378]: "2026"
              - link "Toy Story 5 7.4 Toy Story 5 2026" [ref=e379] [cursor=pointer]:
                - /url: /movie/1084244
                - generic [ref=e380]:
                  - img "Toy Story 5" [ref=e381]
                  - generic [ref=e384]:
                    - img [ref=e385]
                    - generic [ref=e387]: "7.4"
                - heading "Toy Story 5" [level=3] [ref=e388]
                - paragraph [ref=e389]: "2026"
              - link "Bhooth Bangla 5.6 Bhooth Bangla 2026" [ref=e390] [cursor=pointer]:
                - /url: /movie/1239134
                - generic [ref=e391]:
                  - img "Bhooth Bangla" [ref=e392]
                  - generic [ref=e395]:
                    - img [ref=e396]
                    - generic [ref=e398]: "5.6"
                - heading "Bhooth Bangla" [level=3] [ref=e399]
                - paragraph [ref=e400]: "2026"
              - link "Deep Water 6.1 Deep Water 2026" [ref=e401] [cursor=pointer]:
                - /url: /movie/1127384
                - generic [ref=e402]:
                  - img "Deep Water" [ref=e403]
                  - generic [ref=e406]:
                    - img [ref=e407]
                    - generic [ref=e409]: "6.1"
                - heading "Deep Water" [level=3] [ref=e410]
                - paragraph [ref=e411]: "2026"
              - link "The Mandalorian and Grogu 6.7 The Mandalorian and Grogu 2026" [ref=e412] [cursor=pointer]:
                - /url: /movie/1228710
                - generic [ref=e413]:
                  - img "The Mandalorian and Grogu" [ref=e414]
                  - generic [ref=e417]:
                    - img [ref=e418]
                    - generic [ref=e420]: "6.7"
                - heading "The Mandalorian and Grogu" [level=3] [ref=e421]
                - paragraph [ref=e422]: "2026"
              - 'link "Your Fault: London 7.5 Your Fault: London 2026" [ref=e423] [cursor=pointer]':
                - /url: /movie/1477317
                - generic [ref=e424]:
                  - 'img "Your Fault: London" [ref=e425]'
                  - generic [ref=e428]:
                    - img [ref=e429]
                    - generic [ref=e431]: "7.5"
                - 'heading "Your Fault: London" [level=3] [ref=e432]'
                - paragraph [ref=e433]: "2026"
              - link "Disclosure Day 6.7 Disclosure Day 2026" [ref=e434] [cursor=pointer]:
                - /url: /movie/1275779
                - generic [ref=e435]:
                  - img "Disclosure Day" [ref=e436]
                  - generic [ref=e439]:
                    - img [ref=e440]
                    - generic [ref=e442]: "6.7"
                - heading "Disclosure Day" [level=3] [ref=e443]
                - paragraph [ref=e444]: "2026"
          - generic [ref=e445]:
            - generic [ref=e446]:
              - img [ref=e448]
              - heading "Upcoming Movies" [level=2] [ref=e450]
              - img [ref=e451]
            - generic [ref=e453]:
              - link "Obsession 7.9 Obsession 2026" [ref=e454] [cursor=pointer]:
                - /url: /movie/1339713
                - generic [ref=e455]:
                  - img "Obsession" [ref=e456]
                  - generic [ref=e459]:
                    - img [ref=e460]
                    - generic [ref=e462]: "7.9"
                - heading "Obsession" [level=3] [ref=e463]
                - paragraph [ref=e464]: "2026"
              - link "Toy Story 5 7.4 Toy Story 5 2026" [ref=e465] [cursor=pointer]:
                - /url: /movie/1084244
                - generic [ref=e466]:
                  - img "Toy Story 5" [ref=e467]
                  - generic [ref=e470]:
                    - img [ref=e471]
                    - generic [ref=e473]: "7.4"
                - heading "Toy Story 5" [level=3] [ref=e474]
                - paragraph [ref=e475]: "2026"
              - link "Deep Water 6.1 Deep Water 2026" [ref=e476] [cursor=pointer]:
                - /url: /movie/1127384
                - generic [ref=e477]:
                  - img "Deep Water" [ref=e478]
                  - generic [ref=e481]:
                    - img [ref=e482]
                    - generic [ref=e484]: "6.1"
                - heading "Deep Water" [level=3] [ref=e485]
                - paragraph [ref=e486]: "2026"
              - link "The Furious 7.7 The Furious 2026" [ref=e487] [cursor=pointer]:
                - /url: /movie/1280738
                - generic [ref=e488]:
                  - img "The Furious" [ref=e489]
                  - generic [ref=e492]:
                    - img [ref=e493]
                    - generic [ref=e495]: "7.7"
                - heading "The Furious" [level=3] [ref=e496]
                - paragraph [ref=e497]: "2026"
              - link "Supergirl 6.2 Supergirl 2026" [ref=e498] [cursor=pointer]:
                - /url: /movie/1081003
                - generic [ref=e499]:
                  - img "Supergirl" [ref=e500]
                  - generic [ref=e503]:
                    - img [ref=e504]
                    - generic [ref=e506]: "6.2"
                - heading "Supergirl" [level=3] [ref=e507]
                - paragraph [ref=e508]: "2026"
              - 'link "Dhurandhar: The Revenge 7.3 Dhurandhar: The Revenge 2026" [ref=e509] [cursor=pointer]':
                - /url: /movie/1582770
                - generic [ref=e510]:
                  - 'img "Dhurandhar: The Revenge" [ref=e511]'
                  - generic [ref=e514]:
                    - img [ref=e515]
                    - generic [ref=e517]: "7.3"
                - 'heading "Dhurandhar: The Revenge" [level=3] [ref=e518]'
                - paragraph [ref=e519]: "2026"
              - link "Hokum 6.9 Hokum 2026" [ref=e520] [cursor=pointer]:
                - /url: /movie/1430077
                - generic [ref=e521]:
                  - img "Hokum" [ref=e522]
                  - generic [ref=e525]:
                    - img [ref=e526]
                    - generic [ref=e528]: "6.9"
                - heading "Hokum" [level=3] [ref=e529]
                - paragraph [ref=e530]: "2026"
              - link "Tuner 7.2 Tuner 2026" [ref=e531] [cursor=pointer]:
                - /url: /movie/1340206
                - generic [ref=e532]:
                  - img "Tuner" [ref=e533]
                  - generic [ref=e536]:
                    - img [ref=e537]
                    - generic [ref=e539]: "7.2"
                - heading "Tuner" [level=3] [ref=e540]
                - paragraph [ref=e541]: "2026"
              - link "Scary Movie 5.5 Scary Movie 2026" [ref=e542] [cursor=pointer]:
                - /url: /movie/1273221
                - generic [ref=e543]:
                  - img "Scary Movie" [ref=e544]
                  - generic [ref=e547]:
                    - img [ref=e548]
                    - generic [ref=e550]: "5.5"
                - heading "Scary Movie" [level=3] [ref=e551]
                - paragraph [ref=e552]: "2026"
              - 'link "Greenland 2: Migration 6.4 Greenland 2: Migration 2026" [ref=e553] [cursor=pointer]':
                - /url: /movie/840464
                - generic [ref=e554]:
                  - 'img "Greenland 2: Migration" [ref=e555]'
                  - generic [ref=e558]:
                    - img [ref=e559]
                    - generic [ref=e561]: "6.4"
                - 'heading "Greenland 2: Migration" [level=3] [ref=e562]'
                - paragraph [ref=e563]: "2026"
          - generic [ref=e564]:
            - generic [ref=e565]:
              - img [ref=e567]
              - heading "Top Rated Movies" [level=2] [ref=e569]
              - img [ref=e570]
            - generic [ref=e572]:
              - link "Swapped 9.0 Swapped 2026" [ref=e573] [cursor=pointer]:
                - /url: /movie/1007757
                - generic [ref=e574]:
                  - img "Swapped" [ref=e575]
                  - generic [ref=e578]:
                    - img [ref=e579]
                    - generic [ref=e581]: "9.0"
                - heading "Swapped" [level=3] [ref=e582]
                - paragraph [ref=e583]: "2026"
              - link "The Shawshank Redemption 8.7 The Shawshank Redemption 1994" [ref=e584] [cursor=pointer]:
                - /url: /movie/278
                - generic [ref=e585]:
                  - img "The Shawshank Redemption" [ref=e586]
                  - generic [ref=e589]:
                    - img [ref=e590]
                    - generic [ref=e592]: "8.7"
                - heading "The Shawshank Redemption" [level=3] [ref=e593]
                - paragraph [ref=e594]: "1994"
              - link "Project Hail Mary 8.7 Project Hail Mary 2026" [ref=e595] [cursor=pointer]:
                - /url: /movie/687163
                - generic [ref=e596]:
                  - img "Project Hail Mary" [ref=e597]
                  - generic [ref=e600]:
                    - img [ref=e601]
                    - generic [ref=e603]: "8.7"
                - heading "Project Hail Mary" [level=3] [ref=e604]
                - paragraph [ref=e605]: "2026"
              - link "Michael 8.7 Michael 2026" [ref=e606] [cursor=pointer]:
                - /url: /movie/936075
                - generic [ref=e607]:
                  - img "Michael" [ref=e608]
                  - generic [ref=e611]:
                    - img [ref=e612]
                    - generic [ref=e614]: "8.7"
                - heading "Michael" [level=3] [ref=e615]
                - paragraph [ref=e616]: "2026"
              - link "The Godfather 8.7 The Godfather 1972" [ref=e617] [cursor=pointer]:
                - /url: /movie/238
                - generic [ref=e618]:
                  - img "The Godfather" [ref=e619]
                  - generic [ref=e622]:
                    - img [ref=e623]
                    - generic [ref=e625]: "8.7"
                - heading "The Godfather" [level=3] [ref=e626]
                - paragraph [ref=e627]: "1972"
              - link "Remarkably Bright Creatures 8.6 Remarkably Bright Creatures 2026" [ref=e628] [cursor=pointer]:
                - /url: /movie/1330021
                - generic [ref=e629]:
                  - img "Remarkably Bright Creatures" [ref=e630]
                  - generic [ref=e633]:
                    - img [ref=e634]
                    - generic [ref=e636]: "8.6"
                - heading "Remarkably Bright Creatures" [level=3] [ref=e637]
                - paragraph [ref=e638]: "2026"
              - link "The Godfather Part II 8.6 The Godfather Part II 1974" [ref=e639] [cursor=pointer]:
                - /url: /movie/240
                - generic [ref=e640]:
                  - img "The Godfather Part II" [ref=e641]
                  - generic [ref=e644]:
                    - img [ref=e645]
                    - generic [ref=e647]: "8.6"
                - heading "The Godfather Part II" [level=3] [ref=e648]
                - paragraph [ref=e649]: "1974"
              - link "Schindler's List 8.6 Schindler's List 1993" [ref=e650] [cursor=pointer]:
                - /url: /movie/424
                - generic [ref=e651]:
                  - img "Schindler's List" [ref=e652]
                  - generic [ref=e655]:
                    - img [ref=e656]
                    - generic [ref=e658]: "8.6"
                - heading "Schindler's List" [level=3] [ref=e659]
                - paragraph [ref=e660]: "1993"
              - link "12 Angry Men 8.6 12 Angry Men 1957" [ref=e661] [cursor=pointer]:
                - /url: /movie/389
                - generic [ref=e662]:
                  - img "12 Angry Men" [ref=e663]
                  - generic [ref=e666]:
                    - img [ref=e667]
                    - generic [ref=e669]: "8.6"
                - heading "12 Angry Men" [level=3] [ref=e670]
                - paragraph [ref=e671]: "1957"
              - link "Spirited Away 8.5 Spirited Away 2001" [ref=e672] [cursor=pointer]:
                - /url: /movie/129
                - generic [ref=e673]:
                  - img "Spirited Away" [ref=e674]
                  - generic [ref=e677]:
                    - img [ref=e678]
                    - generic [ref=e680]: "8.5"
                - heading "Spirited Away" [level=3] [ref=e681]
                - paragraph [ref=e682]: "2001"
          - generic [ref=e683]:
            - generic [ref=e684]:
              - img [ref=e686]
              - heading "Trending TV Shows" [level=2] [ref=e689]
              - img [ref=e690]
            - generic [ref=e692]:
              - link "House of the Dragon 8.3 House of the Dragon 2022" [ref=e693] [cursor=pointer]:
                - /url: /tv/94997
                - generic [ref=e694]:
                  - img "House of the Dragon" [ref=e695]
                  - generic [ref=e698]:
                    - img [ref=e699]
                    - generic [ref=e701]: "8.3"
                - heading "House of the Dragon" [level=3] [ref=e702]
                - paragraph [ref=e703]: "2022"
              - link "The Bear 8.2 The Bear 2022" [ref=e704] [cursor=pointer]:
                - /url: /tv/136315
                - generic [ref=e705]:
                  - img "The Bear" [ref=e706]
                  - generic [ref=e709]:
                    - img [ref=e710]
                    - generic [ref=e712]: "8.2"
                - heading "The Bear" [level=3] [ref=e713]
                - paragraph [ref=e714]: "2022"
              - link "I Will Find You 8.4 I Will Find You 2026" [ref=e715] [cursor=pointer]:
                - /url: /tv/278178
                - generic [ref=e716]:
                  - img "I Will Find You" [ref=e717]
                  - generic [ref=e720]:
                    - img [ref=e721]
                    - generic [ref=e723]: "8.4"
                - heading "I Will Find You" [level=3] [ref=e724]
                - paragraph [ref=e725]: "2026"
              - link "FROM 8.5 FROM 2022" [ref=e726] [cursor=pointer]:
                - /url: /tv/124364
                - generic [ref=e727]:
                  - img "FROM" [ref=e728]
                  - generic [ref=e731]:
                    - img [ref=e732]
                    - generic [ref=e734]: "8.5"
                - heading "FROM" [level=3] [ref=e735]
                - paragraph [ref=e736]: "2022"
              - link "The Agency 7.1 The Agency 2024" [ref=e737] [cursor=pointer]:
                - /url: /tv/219971
                - generic [ref=e738]:
                  - img "The Agency" [ref=e739]
                  - generic [ref=e742]:
                    - img [ref=e743]
                    - generic [ref=e745]: "7.1"
                - heading "The Agency" [level=3] [ref=e746]
                - paragraph [ref=e747]: "2024"
              - link "Widow's Bay 8.2 Widow's Bay 2026" [ref=e748] [cursor=pointer]:
                - /url: /tv/270476
                - generic [ref=e749]:
                  - img "Widow's Bay" [ref=e750]
                  - generic [ref=e753]:
                    - img [ref=e754]
                    - generic [ref=e756]: "8.2"
                - heading "Widow's Bay" [level=3] [ref=e757]
                - paragraph [ref=e758]: "2026"
              - link "Re:ZERO -Starting Life in Another World- 7.9 Re:ZERO -Starting Life in Another World- 2016" [ref=e759] [cursor=pointer]:
                - /url: /tv/65942
                - generic [ref=e760]:
                  - img "Re:ZERO -Starting Life in Another World-" [ref=e761]
                  - generic [ref=e764]:
                    - img [ref=e765]
                    - generic [ref=e767]: "7.9"
                - heading "Re:ZERO -Starting Life in Another World-" [level=3] [ref=e768]
                - paragraph [ref=e769]: "2016"
              - link "One Piece 8.7 One Piece 1999" [ref=e770] [cursor=pointer]:
                - /url: /tv/37854
                - generic [ref=e771]:
                  - img "One Piece" [ref=e772]
                  - generic [ref=e775]:
                    - img [ref=e776]
                    - generic [ref=e778]: "8.7"
                - heading "One Piece" [level=3] [ref=e779]
                - paragraph [ref=e780]: "1999"
              - link "Rick and Morty 8.7 Rick and Morty 2013" [ref=e781] [cursor=pointer]:
                - /url: /tv/60625
                - generic [ref=e782]:
                  - img "Rick and Morty" [ref=e783]
                  - generic [ref=e786]:
                    - img [ref=e787]
                    - generic [ref=e789]: "8.7"
                - heading "Rick and Morty" [level=3] [ref=e790]
                - paragraph [ref=e791]: "2013"
              - 'link "Avatar: The Last Airbender 7.8 Avatar: The Last Airbender 2024" [ref=e792] [cursor=pointer]':
                - /url: /tv/82452
                - generic [ref=e793]:
                  - 'img "Avatar: The Last Airbender" [ref=e794]'
                  - generic [ref=e797]:
                    - img [ref=e798]
                    - generic [ref=e800]: "7.8"
                - 'heading "Avatar: The Last Airbender" [level=3] [ref=e801]'
                - paragraph [ref=e802]: "2024"
          - generic [ref=e803]:
            - generic [ref=e804]:
              - img [ref=e806]
              - heading "Popular TV Shows" [level=2] [ref=e809]
              - img [ref=e810]
            - generic [ref=e812]:
              - link "FROM 8.5 FROM 2022" [ref=e813] [cursor=pointer]:
                - /url: /tv/124364
                - generic [ref=e814]:
                  - img "FROM" [ref=e815]
                  - generic [ref=e818]:
                    - img [ref=e819]
                    - generic [ref=e821]: "8.5"
                - heading "FROM" [level=3] [ref=e822]
                - paragraph [ref=e823]: "2022"
              - link "House of the Dragon 8.3 House of the Dragon 2022" [ref=e824] [cursor=pointer]:
                - /url: /tv/94997
                - generic [ref=e825]:
                  - img "House of the Dragon" [ref=e826]
                  - generic [ref=e829]:
                    - img [ref=e830]
                    - generic [ref=e832]: "8.3"
                - heading "House of the Dragon" [level=3] [ref=e833]
                - paragraph [ref=e834]: "2022"
              - link "Raakh 9.0 Raakh 2026" [ref=e835] [cursor=pointer]:
                - /url: /tv/284631
                - generic [ref=e836]:
                  - img "Raakh" [ref=e837]
                  - generic [ref=e840]:
                    - img [ref=e841]
                    - generic [ref=e843]: "9.0"
                - heading "Raakh" [level=3] [ref=e844]
                - paragraph [ref=e845]: "2026"
              - link "Off Campus 8.9 Off Campus 2026" [ref=e846] [cursor=pointer]:
                - /url: /tv/273240
                - generic [ref=e847]:
                  - img "Off Campus" [ref=e848]
                  - generic [ref=e851]:
                    - img [ref=e852]
                    - generic [ref=e854]: "8.9"
                - heading "Off Campus" [level=3] [ref=e855]
                - paragraph [ref=e856]: "2026"
              - link "Wanna Be Happy? 7.2 Wanna Be Happy? 2021" [ref=e857] [cursor=pointer]:
                - /url: /tv/107447
                - generic [ref=e858]:
                  - img "Wanna Be Happy?" [ref=e859]
                  - generic [ref=e862]:
                    - img [ref=e863]
                    - generic [ref=e865]: "7.2"
                - heading "Wanna Be Happy?" [level=3] [ref=e866]
                - paragraph [ref=e867]: "2021"
              - 'link "Law & Order: Special Victims Unit 7.9 Law & Order: Special Victims Unit 1999" [ref=e868] [cursor=pointer]':
                - /url: /tv/2734
                - generic [ref=e869]:
                  - 'img "Law & Order: Special Victims Unit" [ref=e870]'
                  - generic [ref=e873]:
                    - img [ref=e874]
                    - generic [ref=e876]: "7.9"
                - 'heading "Law & Order: Special Victims Unit" [level=3] [ref=e877]'
                - paragraph [ref=e878]: "1999"
              - link "The Boys 8.4 The Boys 2019" [ref=e879] [cursor=pointer]:
                - /url: /tv/76479
                - generic [ref=e880]:
                  - img "The Boys" [ref=e881]
                  - generic [ref=e884]:
                    - img [ref=e885]
                    - generic [ref=e887]: "8.4"
                - heading "The Boys" [level=3] [ref=e888]
                - paragraph [ref=e889]: "2019"
              - link "The Rookie 8.6 The Rookie 2018" [ref=e890] [cursor=pointer]:
                - /url: /tv/79744
                - generic [ref=e891]:
                  - img "The Rookie" [ref=e892]
                  - generic [ref=e895]:
                    - img [ref=e896]
                    - generic [ref=e898]: "8.6"
                - heading "The Rookie" [level=3] [ref=e899]
                - paragraph [ref=e900]: "2018"
              - link "Law & Order 7.3 Law & Order 1990" [ref=e901] [cursor=pointer]:
                - /url: /tv/549
                - generic [ref=e902]:
                  - img "Law & Order" [ref=e903]
                  - generic [ref=e906]:
                    - img [ref=e907]
                    - generic [ref=e909]: "7.3"
                - heading "Law & Order" [level=3] [ref=e910]
                - paragraph [ref=e911]: "1990"
              - link "Game of Thrones 8.5 Game of Thrones 2011" [ref=e912] [cursor=pointer]:
                - /url: /tv/1399
                - generic [ref=e913]:
                  - img "Game of Thrones" [ref=e914]
                  - generic [ref=e917]:
                    - img [ref=e918]
                    - generic [ref=e920]: "8.5"
                - heading "Game of Thrones" [level=3] [ref=e921]
                - paragraph [ref=e922]: "2011"
          - generic [ref=e923]:
            - generic [ref=e924]:
              - img [ref=e926]
              - heading "Airing Today" [level=2] [ref=e929]
              - img [ref=e930]
            - generic [ref=e932]:
              - link "Never-Ending Summer 8.8 Never-Ending Summer 2026" [ref=e933] [cursor=pointer]:
                - /url: /tv/288603
                - generic [ref=e934]:
                  - img "Never-Ending Summer" [ref=e935]
                  - generic [ref=e938]:
                    - img [ref=e939]
                    - generic [ref=e941]: "8.8"
                - heading "Never-Ending Summer" [level=3] [ref=e942]
                - paragraph [ref=e943]: "2026"
              - link "Criminal Minds 8.3 Criminal Minds 2005" [ref=e944] [cursor=pointer]:
                - /url: /tv/4057
                - generic [ref=e945]:
                  - img "Criminal Minds" [ref=e946]
                  - generic [ref=e949]:
                    - img [ref=e950]
                    - generic [ref=e952]: "8.3"
                - heading "Criminal Minds" [level=3] [ref=e953]
                - paragraph [ref=e954]: "2005"
              - link "Watch What Happens Live with Andy Cohen 5.0 Watch What Happens Live with Andy Cohen 2009" [ref=e955] [cursor=pointer]:
                - /url: /tv/22980
                - generic [ref=e956]:
                  - img "Watch What Happens Live with Andy Cohen" [ref=e957]
                  - generic [ref=e960]:
                    - img [ref=e961]
                    - generic [ref=e963]: "5.0"
                - heading "Watch What Happens Live with Andy Cohen" [level=3] [ref=e964]
                - paragraph [ref=e965]: "2009"
              - link "Binnelanders 5.5 Binnelanders 2005" [ref=e966] [cursor=pointer]:
                - /url: /tv/206559
                - generic [ref=e967]:
                  - img "Binnelanders" [ref=e968]
                  - generic [ref=e971]:
                    - img [ref=e972]
                    - generic [ref=e974]: "5.5"
                - heading "Binnelanders" [level=3] [ref=e975]
                - paragraph [ref=e976]: "2005"
              - link "Coronation Street 5.3 Coronation Street 1960" [ref=e977] [cursor=pointer]:
                - /url: /tv/291
                - generic [ref=e978]:
                  - img "Coronation Street" [ref=e979]
                  - generic [ref=e982]:
                    - img [ref=e983]
                    - generic [ref=e985]: "5.3"
                - heading "Coronation Street" [level=3] [ref=e986]
                - paragraph [ref=e987]: "1960"
              - link "Wonderland 6.8 Wonderland 2021" [ref=e988] [cursor=pointer]:
                - /url: /tv/126398
                - generic [ref=e989]:
                  - img "Wonderland" [ref=e990]
                  - generic [ref=e993]:
                    - img [ref=e994]
                    - generic [ref=e996]: "6.8"
                - heading "Wonderland" [level=3] [ref=e997]
                - paragraph [ref=e998]: "2021"
              - link "Jeopardy! 6.9 Jeopardy! 1984" [ref=e999] [cursor=pointer]:
                - /url: /tv/2912
                - generic [ref=e1000]:
                  - img "Jeopardy!" [ref=e1001]
                  - generic [ref=e1004]:
                    - img [ref=e1005]
                    - generic [ref=e1007]: "6.9"
                - heading "Jeopardy!" [level=3] [ref=e1008]
                - paragraph [ref=e1009]: "1984"
              - link "Gute Zeiten, schlechte Zeiten 5.5 Gute Zeiten, schlechte Zeiten 1992" [ref=e1010] [cursor=pointer]:
                - /url: /tv/13945
                - generic [ref=e1011]:
                  - img "Gute Zeiten, schlechte Zeiten" [ref=e1012]
                  - generic [ref=e1015]:
                    - img [ref=e1016]
                    - generic [ref=e1018]: "5.5"
                - heading "Gute Zeiten, schlechte Zeiten" [level=3] [ref=e1019]
                - paragraph [ref=e1020]: "1992"
              - link "Cape Fear 6.7 Cape Fear 2026" [ref=e1021] [cursor=pointer]:
                - /url: /tv/277439
                - generic [ref=e1022]:
                  - img "Cape Fear" [ref=e1023]
                  - generic [ref=e1026]:
                    - img [ref=e1027]
                    - generic [ref=e1029]: "6.7"
                - heading "Cape Fear" [level=3] [ref=e1030]
                - paragraph [ref=e1031]: "2026"
              - link "Sugar 7.2 Sugar 2024" [ref=e1032] [cursor=pointer]:
                - /url: /tv/203744
                - generic [ref=e1033]:
                  - img "Sugar" [ref=e1034]
                  - generic [ref=e1037]:
                    - img [ref=e1038]
                    - generic [ref=e1040]: "7.2"
                - heading "Sugar" [level=3] [ref=e1041]
                - paragraph [ref=e1042]: "2024"
          - generic [ref=e1043]:
            - generic [ref=e1044]:
              - img [ref=e1046]
              - heading "Top Rated TV" [level=2] [ref=e1048]
              - img [ref=e1049]
            - generic [ref=e1051]:
              - link "Teach You a Lesson 9.5 Teach You a Lesson 2026" [ref=e1052] [cursor=pointer]:
                - /url: /tv/276161
                - generic [ref=e1053]:
                  - img "Teach You a Lesson" [ref=e1054]
                  - generic [ref=e1057]:
                    - img [ref=e1058]
                    - generic [ref=e1060]: "9.5"
                - heading "Teach You a Lesson" [level=3] [ref=e1061]
                - paragraph [ref=e1062]: "2026"
              - link "Dutton Ranch 9.3 Dutton Ranch 2026" [ref=e1063] [cursor=pointer]:
                - /url: /tv/299167
                - generic [ref=e1064]:
                  - img "Dutton Ranch" [ref=e1065]
                  - generic [ref=e1068]:
                    - img [ref=e1069]
                    - generic [ref=e1071]: "9.3"
                - heading "Dutton Ranch" [level=3] [ref=e1072]
                - paragraph [ref=e1073]: "2026"
              - link "Breaking Bad 8.9 Breaking Bad 2008" [ref=e1074] [cursor=pointer]:
                - /url: /tv/1396
                - generic [ref=e1075]:
                  - img "Breaking Bad" [ref=e1076]
                  - generic [ref=e1079]:
                    - img [ref=e1080]
                    - generic [ref=e1082]: "8.9"
                - heading "Breaking Bad" [level=3] [ref=e1083]
                - paragraph [ref=e1084]: "2008"
              - link "Off Campus 8.9 Off Campus 2026" [ref=e1085] [cursor=pointer]:
                - /url: /tv/273240
                - generic [ref=e1086]:
                  - img "Off Campus" [ref=e1087]
                  - generic [ref=e1090]:
                    - img [ref=e1091]
                    - generic [ref=e1093]: "8.9"
                - heading "Off Campus" [level=3] [ref=e1094]
                - paragraph [ref=e1095]: "2026"
              - 'link "Frieren: Beyond Journey''s End 8.8 Frieren: Beyond Journey''s End 2023" [ref=e1096] [cursor=pointer]':
                - /url: /tv/209867
                - generic [ref=e1097]:
                  - 'img "Frieren: Beyond Journey''s End" [ref=e1098]'
                  - generic [ref=e1101]:
                    - img [ref=e1102]
                    - generic [ref=e1104]: "8.8"
                - 'heading "Frieren: Beyond Journey''s End" [level=3] [ref=e1105]'
                - paragraph [ref=e1106]: "2023"
              - 'link "Avatar: The Last Airbender 8.8 Avatar: The Last Airbender 2005" [ref=e1107] [cursor=pointer]':
                - /url: /tv/246
                - generic [ref=e1108]:
                  - 'img "Avatar: The Last Airbender" [ref=e1109]'
                  - generic [ref=e1112]:
                    - img [ref=e1113]
                    - generic [ref=e1115]: "8.8"
                - 'heading "Avatar: The Last Airbender" [level=3] [ref=e1116]'
                - paragraph [ref=e1117]: "2005"
              - link "When Life Gives You Tangerines 8.8 When Life Gives You Tangerines 2025" [ref=e1118] [cursor=pointer]:
                - /url: /tv/219246
                - generic [ref=e1119]:
                  - img "When Life Gives You Tangerines" [ref=e1120]
                  - generic [ref=e1123]:
                    - img [ref=e1124]
                    - generic [ref=e1126]: "8.8"
                - heading "When Life Gives You Tangerines" [level=3] [ref=e1127]
                - paragraph [ref=e1128]: "2025"
              - link "Arcane 8.8 Arcane 2021" [ref=e1129] [cursor=pointer]:
                - /url: /tv/94605
                - generic [ref=e1130]:
                  - img "Arcane" [ref=e1131]
                  - generic [ref=e1134]:
                    - img [ref=e1135]
                    - generic [ref=e1137]: "8.8"
                - heading "Arcane" [level=3] [ref=e1138]
                - paragraph [ref=e1139]: "2021"
              - link "The Chosen 8.8 The Chosen 2019" [ref=e1140] [cursor=pointer]:
                - /url: /tv/85077
                - generic [ref=e1141]:
                  - img "The Chosen" [ref=e1142]
                  - generic [ref=e1145]:
                    - img [ref=e1146]
                    - generic [ref=e1148]: "8.8"
                - heading "The Chosen" [level=3] [ref=e1149]
                - paragraph [ref=e1150]: "2019"
              - link "One Piece 8.7 One Piece 1999" [ref=e1151] [cursor=pointer]:
                - /url: /tv/37854
                - generic [ref=e1152]:
                  - img "One Piece" [ref=e1153]
                  - generic [ref=e1156]:
                    - img [ref=e1157]
                    - generic [ref=e1159]: "8.7"
                - heading "One Piece" [level=3] [ref=e1160]
                - paragraph [ref=e1161]: "1999"
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test'
  2   | 
  3   | const BASE_URL = process.env.BASE_URL || 'http://localhost:5178'
  4   | 
  5   | // Helper to wait for API response
  6   | async function waitForAPIResponse(page: import('@playwright/test').Page, urlPattern: string) {
  7   |   return page.waitForResponse(response => response.url().includes(urlPattern))
  8   | }
  9   | 
  10  | test.describe('Media Manager - Auth', () => {
  11  |   test('should show login page', async ({ page }) => {
  12  |     await page.goto(`${BASE_URL}/login`)
  13  |     
  14  |     // Check login form elements
  15  |     await expect(page.locator('h1:has-text("Media Manager")')).toBeVisible()
  16  |     await expect(page.locator('input[placeholder="Enter your username"]')).toBeVisible()
  17  |     await expect(page.locator('input[placeholder="Enter your password"]')).toBeVisible()
  18  |     await expect(page.locator('button:has-text("Sign In")')).toBeVisible()
  19  |   })
  20  | 
  21  |   test('should toggle between login and register', async ({ page }) => {
  22  |     await page.goto(`${BASE_URL}/login`)
  23  |     
  24  |     // Click to register
  25  |     await page.click('text=Don\'t have an account? Create one')
  26  |     
  27  |     await expect(page.locator('h2:has-text("Create Account")')).toBeVisible()
  28  |     await expect(page.locator('button:has-text("Create Account")')).toBeVisible()
  29  |     
  30  |     // Click back to login
  31  |     await page.click('text=Already have an account? Sign in')
  32  |     
  33  |     await expect(page.locator('h2:has-text("Sign In")')).toBeVisible()
  34  |   })
  35  | })
  36  | 
  37  | test.describe('Media Manager - Navigation', () => {
  38  |   test('should navigate to all main pages', async ({ page }) => {
  39  |     await page.goto(BASE_URL)
  40  |     
  41  |     // Check sidebar navigation - includes Watchlist now
  42  |     const navItems = ['Home', 'Discover', 'Watchlist', 'Downloads', 'Library', 'Search', 'Suggestions', 'Settings']
  43  |     
  44  |     for (const item of navItems) {
  45  |       await page.click(`text=${item}`)
  46  |       await page.waitForLoadState('networkidle')
  47  |       
  48  |       // Verify we're on the right page by checking URL
  49  |       const url = page.url()
  50  |       expect(url).toContain(item.toLowerCase())
  51  |     }
  52  |   })
  53  | })
  54  | 
  55  | test.describe('Media Manager - Discover Page', () => {
  56  |   test('should load discover page with content', async ({ page }) => {
  57  |     await page.goto(`${BASE_URL}/discover`)
  58  |     
  59  |     // Wait for API responses
  60  |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  61  |     
  62  |     // Check page title
  63  |     await expect(page.locator('h1:has-text("Discover")')).toBeVisible()
  64  |     
  65  |     // Check tabs
  66  |     await expect(page.locator('button:has-text("All")')).toBeVisible()
  67  |     await expect(page.locator('button:has-text("MOVIES")')).toBeVisible()
  68  |     await expect(page.locator('button:has-text("TV SHOWS")')).toBeVisible()
  69  |     
  70  |     // Check content sections
> 71  |     await expect(page.locator('text=Trending Movies')).toBeVisible()
      |                                                        ^ Error: expect(locator).toBeVisible() failed
  72  |     await expect(page.locator('text=Popular Movies')).toBeVisible()
  73  |   })
  74  | 
  75  |   test('should filter discover by movies tab', async ({ page }) => {
  76  |     await page.goto(`${BASE_URL}/discover`)
  77  |     
  78  |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  79  |     
  80  |     // Click Movies tab
  81  |     await page.click('button:has-text("MOVIES")')
  82  |     
  83  |     await page.waitForTimeout(500)
  84  |     
  85  |     // Should show movie sections only
  86  |     await expect(page.locator('text=Trending Movies')).toBeVisible()
  87  |     await expect(page.locator('text=Popular Movies')).toBeVisible()
  88  |     
  89  |     // Should NOT show TV sections
  90  |     const tvContent = await page.locator('text=Trending TV Shows').count()
  91  |     expect(tvContent).toBe(0)
  92  |   })
  93  | 
  94  |   test('should filter discover by TV tab', async ({ page }) => {
  95  |     await page.goto(`${BASE_URL}/discover`)
  96  |     
  97  |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  98  |     
  99  |     // Click TV Shows tab
  100 |     await page.click('button:has-text("TV SHOWS")')
  101 |     
  102 |     await page.waitForTimeout(500)
  103 |     
  104 |     // Should show TV sections
  105 |     await expect(page.locator('text=Trending TV Shows')).toBeVisible()
  106 |     await expect(page.locator('text=Popular TV Shows')).toBeVisible()
  107 |   })
  108 | 
  109 |   test('should show movie cards with ratings', async ({ page }) => {
  110 |     await page.goto(`${BASE_URL}/discover`)
  111 |     
  112 |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  113 |     
  114 |     // Check for movie cards with star ratings
  115 |     const starRatings = page.locator('.text-yellow-400')
  116 |     await expect(starRatings.first()).toBeVisible()
  117 |     
  118 |     // Check for movie titles
  119 |     const movieTitles = page.locator('h3.text-white')
  120 |     await expect(movieTitles.first()).toBeVisible()
  121 |   })
  122 | })
  123 | 
  124 | test.describe('Media Manager - Movie Detail', () => {
  125 |   test('should navigate to movie detail page', async ({ page }) => {
  126 |     await page.goto(`${BASE_URL}/discover`)
  127 |     
  128 |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  129 |     
  130 |     // Click first movie card
  131 |     await page.click('a[href^="/movie/"]')
  132 |     
  133 |     // Wait for movie detail API
  134 |     await waitForAPIResponse(page, '/api/discover/movie/')
  135 |     
  136 |     // Check movie detail elements
  137 |     await expect(page.locator('h1')).toBeVisible()
  138 |     await expect(page.locator('text=Overview')).toBeVisible()
  139 |     await expect(page.locator('text=Cast')).toBeVisible()
  140 |   })
  141 | 
  142 |   test('should show movie info sidebar', async ({ page }) => {
  143 |     await page.goto(`${BASE_URL}/movie/550`) // Fight Club as example
  144 |     
  145 |     await waitForAPIResponse(page, '/api/discover/movie/')
  146 |     
  147 |     // Check info sidebar
  148 |     await expect(page.locator('h3:has-text("Movie Info")')).toBeVisible()
  149 |     await expect(page.locator('text=Release Date')).toBeVisible()
  150 |     await expect(page.locator('text=Runtime')).toBeVisible()
  151 |     await expect(page.locator('text=Budget')).toBeVisible()
  152 |     await expect(page.locator('text=Revenue')).toBeVisible()
  153 |   })
  154 | 
  155 |   test('should show similar movies', async ({ page }) => {
  156 |     await page.goto(`${BASE_URL}/movie/550`)
  157 |     
  158 |     await waitForAPIResponse(page, '/api/discover/movie/')
  159 |     
  160 |     // Check similar movies section
  161 |     await expect(page.locator('h2:has-text("Similar Movies")')).toBeVisible()
  162 |   })
  163 | })
  164 | 
  165 | test.describe('Media Manager - TV Detail', () => {
  166 |   test('should navigate to TV detail page', async ({ page }) => {
  167 |     await page.goto(`${BASE_URL}/discover`)
  168 |     
  169 |     await waitForAPIResponse(page, '/api/discover/tv/trending')
  170 |     
  171 |     // Click TV Shows tab first
```