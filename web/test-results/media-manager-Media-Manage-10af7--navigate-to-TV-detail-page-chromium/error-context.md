# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: media-manager.spec.ts >> Media Manager - TV Detail >> should navigate to TV detail page
- Location: e2e/media-manager.spec.ts:166:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('text=Episodes')
Expected: visible
Error: strict mode violation: locator('text=Episodes') resolved to 7 elements:
    1) <span class="text-[#b3b3b3]">26 Episodes</span> aka getByText('26 Episodes')
    2) <h2 class="text-2xl font-bold text-white">Episodes</h2> aka getByRole('heading', { name: 'Episodes' })
    3) <option value="0">Specials (73 episodes)</option> aka getByRole('combobox')
    4) <option value="1">Season 1 (10 episodes)</option> aka getByRole('combobox')
    5) <option value="2">Season 2 (8 episodes)</option> aka getByRole('combobox')
    6) <option value="3">Season 3 (8 episodes)</option> aka getByRole('combobox')
    7) <p class="text-[#b3b3b3] text-sm">Episodes</p> aka getByRole('paragraph').filter({ hasText: 'Episodes' })

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('text=Episodes')

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
        - generic [ref=e70]:
          - link [ref=e73] [cursor=pointer]:
            - /url: /discover
            - img [ref=e74]
          - generic [ref=e77]:
            - heading "House of the Dragon" [level=1] [ref=e78]
            - paragraph [ref=e79]: Win or die.
            - generic [ref=e80]:
              - generic [ref=e81]:
                - img [ref=e82]
                - generic [ref=e84]: "8.3"
                - generic [ref=e85]: (6,283 votes)
              - generic [ref=e86]: "|"
              - generic [ref=e87]: "2022"
              - generic [ref=e88]: "|"
              - generic [ref=e89]: 3 Seasons
              - generic [ref=e90]: "|"
              - generic [ref=e91]: 26 Episodes
            - generic [ref=e92]:
              - generic [ref=e93]: Sci-Fi & Fantasy
              - generic [ref=e94]: Drama
              - generic [ref=e95]: Action & Adventure
            - generic [ref=e96]:
              - button "Watch Trailer" [ref=e97] [cursor=pointer]:
                - img [ref=e98]
                - text: Watch Trailer
              - button "Add to Watchlist" [ref=e100] [cursor=pointer]:
                - img [ref=e101]
                - text: Add to Watchlist
              - button "Download" [ref=e103] [cursor=pointer]:
                - img [ref=e104]
                - text: Download
        - generic [ref=e107]:
          - generic [ref=e108]:
            - generic [ref=e109]:
              - generic [ref=e110]:
                - heading "Overview" [level=2] [ref=e111]
                - paragraph [ref=e112]: The Targaryen dynasty is at the absolute apex of its power, with more than 15 dragons under their yoke. Most empires crumble from such heights. In the case of the Targaryens, their slow fall begins when King Viserys breaks with a century of tradition by naming his daughter Rhaenyra heir to the Iron Throne. But when Viserys later fathers a son, the court is shocked when Rhaenyra retains her status as his heir, and seeds of division sow friction across the realm.
              - generic [ref=e113]:
                - generic [ref=e114]:
                  - heading "Episodes" [level=2] [ref=e115]
                  - generic [ref=e116]:
                    - combobox [ref=e117] [cursor=pointer]:
                      - option "Specials (73 episodes)" [selected]
                      - option "Season 1 (10 episodes)"
                      - option "Season 2 (8 episodes)"
                      - option "Season 3 (8 episodes)"
                    - img
                - generic [ref=e118]:
                  - generic [ref=e119] [cursor=pointer]:
                    - img "The House That Dragons Built S1 E1" [ref=e121]
                    - generic [ref=e122]:
                      - generic [ref=e123]:
                        - generic [ref=e124]: S0E1
                        - heading "The House That Dragons Built S1 E1" [level=3] [ref=e125]
                      - paragraph [ref=e126]: A behind-the-scenes look at the logistics, set design, and special effects used in the series premiere of House of the Dragon.
                      - generic [ref=e127]:
                        - generic [ref=e128]: 2022-08-21
                        - generic [ref=e129]: 26 min
                        - generic [ref=e130]:
                          - img [ref=e131]
                          - generic [ref=e133]: "4.6"
                    - button [ref=e135]:
                      - img [ref=e136]
                  - generic [ref=e138] [cursor=pointer]:
                    - img "The House That Dragons Built S1 E2" [ref=e140]
                    - generic [ref=e141]:
                      - generic [ref=e142]:
                        - generic [ref=e143]: S0E2
                        - heading "The House That Dragons Built S1 E2" [level=3] [ref=e144]
                      - paragraph [ref=e145]: A behind-the-scenes look at the crafting of S1 E2 of the HBO original series House of the Dragon.
                      - generic [ref=e146]:
                        - generic [ref=e147]: 2022-08-28
                        - generic [ref=e148]: 22 min
                        - generic [ref=e149]:
                          - img [ref=e150]
                          - generic [ref=e152]: "6.0"
                    - button [ref=e154]:
                      - img [ref=e155]
                  - generic [ref=e157] [cursor=pointer]:
                    - img "The House That Dragons Built S1 E3" [ref=e159]
                    - generic [ref=e160]:
                      - generic [ref=e161]:
                        - generic [ref=e162]: S0E3
                        - heading "The House That Dragons Built S1 E3" [level=3] [ref=e163]
                      - paragraph [ref=e164]: A behind-the-scenes look at the crafting of S1 E3 of the HBO original series House of the Dragon.
                      - generic [ref=e165]:
                        - generic [ref=e166]: 2022-09-04
                        - generic [ref=e167]: 20 min
                        - generic [ref=e168]:
                          - img [ref=e169]
                          - generic [ref=e171]: "4.5"
                    - button [ref=e173]:
                      - img [ref=e174]
                  - generic [ref=e176] [cursor=pointer]:
                    - img "The House That Dragons Built S1 E4" [ref=e178]
                    - generic [ref=e179]:
                      - generic [ref=e180]:
                        - generic [ref=e181]: S0E4
                        - heading "The House That Dragons Built S1 E4" [level=3] [ref=e182]
                      - paragraph [ref=e183]: A behind-the-scenes look at the crafting of S1 E4 of the HBO original series House of the Dragon.
                      - generic [ref=e184]:
                        - generic [ref=e185]: 2022-09-11
                        - generic [ref=e186]: 23 min
                        - generic [ref=e187]:
                          - img [ref=e188]
                          - generic [ref=e190]: "4.5"
                    - button [ref=e192]:
                      - img [ref=e193]
                  - generic [ref=e195] [cursor=pointer]:
                    - img "The House That Dragons Built S1 E5" [ref=e197]
                    - generic [ref=e198]:
                      - generic [ref=e199]:
                        - generic [ref=e200]: S0E5
                        - heading "The House That Dragons Built S1 E5" [level=3] [ref=e201]
                      - paragraph [ref=e202]: A behind-the-scenes look at the crafting of S1 E5 of the HBO original series House of the Dragon.
                      - generic [ref=e203]:
                        - generic [ref=e204]: 2022-09-18
                        - generic [ref=e205]: 24 min
                        - generic [ref=e206]:
                          - img [ref=e207]
                          - generic [ref=e209]: "4.5"
                    - button [ref=e211]:
                      - img [ref=e212]
                  - generic [ref=e214] [cursor=pointer]:
                    - img "The House That Dragons Built S1 E6" [ref=e216]
                    - generic [ref=e217]:
                      - generic [ref=e218]:
                        - generic [ref=e219]: S0E6
                        - heading "The House That Dragons Built S1 E6" [level=3] [ref=e220]
                      - paragraph [ref=e221]: A behind-the-scenes look at the crafting of S1 E6 of the HBO original series House of the Dragon.
                      - generic [ref=e222]:
                        - generic [ref=e223]: 2022-09-25
                        - generic [ref=e224]: 26 min
                        - generic [ref=e225]:
                          - img [ref=e226]
                          - generic [ref=e228]: "4.5"
                    - button [ref=e230]:
                      - img [ref=e231]
                  - generic [ref=e233] [cursor=pointer]:
                    - img "The House That Dragons Built S1 E7" [ref=e235]
                    - generic [ref=e236]:
                      - generic [ref=e237]:
                        - generic [ref=e238]: S0E7
                        - heading "The House That Dragons Built S1 E7" [level=3] [ref=e239]
                      - paragraph [ref=e240]: A behind-the-scenes look at the crafting of S1 E7 of the HBO original series House of the Dragon.
                      - generic [ref=e241]:
                        - generic [ref=e242]: 2022-10-02
                        - generic [ref=e243]: 24 min
                        - generic [ref=e244]:
                          - img [ref=e245]
                          - generic [ref=e247]: "5.7"
                    - button [ref=e249]:
                      - img [ref=e250]
                  - generic [ref=e252] [cursor=pointer]:
                    - img "The House That Dragons Built S1 E8" [ref=e254]
                    - generic [ref=e255]:
                      - generic [ref=e256]:
                        - generic [ref=e257]: S0E8
                        - heading "The House That Dragons Built S1 E8" [level=3] [ref=e258]
                      - paragraph [ref=e259]: A behind-the-scenes look at the crafting of S1 E8 of the HBO original series House of the Dragon.
                      - generic [ref=e260]:
                        - generic [ref=e261]: 2022-10-09
                        - generic [ref=e262]: 23 min
                        - generic [ref=e263]:
                          - img [ref=e264]
                          - generic [ref=e266]: "5.7"
                    - button [ref=e268]:
                      - img [ref=e269]
                  - generic [ref=e271] [cursor=pointer]:
                    - img "The House That Dragons Built S1 E9" [ref=e273]
                    - generic [ref=e274]:
                      - generic [ref=e275]:
                        - generic [ref=e276]: S0E9
                        - heading "The House That Dragons Built S1 E9" [level=3] [ref=e277]
                      - paragraph [ref=e278]: A behind-the-scenes look at the crafting of S1 E9 of the HBO original series House of the Dragon.
                      - generic [ref=e279]:
                        - generic [ref=e280]: 2022-10-16
                        - generic [ref=e281]: 22 min
                        - generic [ref=e282]:
                          - img [ref=e283]
                          - generic [ref=e285]: "5.7"
                    - button [ref=e287]:
                      - img [ref=e288]
                  - generic [ref=e290] [cursor=pointer]:
                    - img "The House That Dragons Built S1 E10" [ref=e292]
                    - generic [ref=e293]:
                      - generic [ref=e294]:
                        - generic [ref=e295]: S0E10
                        - heading "The House That Dragons Built S1 E10" [level=3] [ref=e296]
                      - paragraph [ref=e297]: A behind-the-scenes look at the crafting of S1 E10 of the HBO original series House of the Dragon.
                      - generic [ref=e298]:
                        - generic [ref=e299]: 2022-10-23
                        - generic [ref=e300]: 24 min
                        - generic [ref=e301]:
                          - img [ref=e302]
                          - generic [ref=e304]: "5.7"
                    - button [ref=e306]:
                      - img [ref=e307]
                  - generic [ref=e309] [cursor=pointer]:
                    - img "Inside S1 E1" [ref=e311]
                    - generic [ref=e312]:
                      - generic [ref=e313]:
                        - generic [ref=e314]: S0E11
                        - heading "Inside S1 E1" [level=3] [ref=e315]
                      - paragraph [ref=e316]: Ryan Condal, Miguel Sapochnik, and more discuss why they opened the episode with the Great Council, the juxtaposition between the tournament and Queen Aemma's birthing sequence, and the motivation behind Viserys choosing Rhaenyra as heir.
                      - generic [ref=e317]:
                        - generic [ref=e318]: 2022-08-21
                        - generic [ref=e319]: 7 min
                        - generic [ref=e320]:
                          - img [ref=e321]
                          - generic [ref=e323]: "4.0"
                    - button [ref=e325]:
                      - img [ref=e326]
                  - generic [ref=e328] [cursor=pointer]:
                    - img "Inside S1 E2" [ref=e330]
                    - generic [ref=e331]:
                      - generic [ref=e332]:
                        - generic [ref=e333]: S0E12
                        - heading "Inside S1 E2" [level=3] [ref=e334]
                      - paragraph [ref=e335]: Milly Alcock (Rhaenyra Targaryen), Emily Carey (Alicent Hightower), Ryan Condal, Miguel Sapochnik and more discuss the selection of the Kingsguard, the face-off at Dragonstone, and the pressure Viserys faces to choose a new wife in "The Rogue Prince."
                      - generic [ref=e336]:
                        - generic [ref=e337]: 2022-08-28
                        - generic [ref=e338]: 7 min
                        - generic [ref=e339]:
                          - img [ref=e340]
                          - generic [ref=e342]: "4.0"
                    - button [ref=e344]:
                      - img [ref=e345]
                  - generic [ref=e347] [cursor=pointer]:
                    - img "Inside S1 E3" [ref=e349]
                    - generic [ref=e350]:
                      - generic [ref=e351]:
                        - generic [ref=e352]: S0E13
                        - heading "Inside S1 E3" [level=3] [ref=e353]
                      - paragraph [ref=e354]: Ryan Condal, Miguel Sapochnik, Matt Smith (Daemon Targaryen), and more speak to the timejump in the third episode, Viserys and Rhaenyra's mental states during the hunt, and Daemon's fight at the Stepstones.
                      - generic [ref=e355]:
                        - generic [ref=e356]: 2022-09-04
                        - generic [ref=e357]: 7 min
                        - generic [ref=e358]:
                          - img [ref=e359]
                          - generic [ref=e361]: "4.0"
                    - button [ref=e363]:
                      - img [ref=e364]
                  - generic [ref=e366] [cursor=pointer]:
                    - img "Inside S1 E4" [ref=e368]
                    - generic [ref=e369]:
                      - generic [ref=e370]:
                        - generic [ref=e371]: S0E14
                        - heading "Inside S1 E4" [level=3] [ref=e372]
                      - paragraph [ref=e373]: Ryan Condal, director Clare Kilner, Emily Carey (Alicent Hightower) and more hone in on the motivation behind Daemon's advances towards Rhaenyra, approaching intimate scenes, and the fractures in Rhaenyra and Alicent's friendship.
                      - generic [ref=e374]:
                        - generic [ref=e375]: 2022-09-11
                        - generic [ref=e376]: 7 min
                        - generic [ref=e377]:
                          - img [ref=e378]
                          - generic [ref=e380]: "4.0"
                    - button [ref=e382]:
                      - img [ref=e383]
                  - generic [ref=e385] [cursor=pointer]:
                    - img "Inside S1 E5" [ref=e387]
                    - generic [ref=e388]:
                      - generic [ref=e389]:
                        - generic [ref=e390]: S0E15
                        - heading "Inside S1 E5" [level=3] [ref=e391]
                      - paragraph [ref=e392]: Ryan Condal, director Clare Kilner, Rhys Ifans (Otto Hightower), Emily Carey (Alicent Hightower) and more dive into Rhaenyra and Laenor's marriage arrangement, and where Alicent's allegiances lie.
                      - generic [ref=e393]:
                        - generic [ref=e394]: 2022-09-18
                        - generic [ref=e395]: 7 min
                        - generic [ref=e396]:
                          - img [ref=e397]
                          - generic [ref=e399]: "4.0"
                    - button [ref=e401]:
                      - img [ref=e402]
                  - generic [ref=e404] [cursor=pointer]:
                    - img "Inside S1 E6" [ref=e406]
                    - generic [ref=e407]:
                      - generic [ref=e408]:
                        - generic [ref=e409]: S0E16
                        - heading "Inside S1 E6" [level=3] [ref=e410]
                      - paragraph [ref=e411]: Cast and crew, including Emma D'Arcy (Rhaenyra) and Matt Smith (Daemon), examine how Rhaenyra's lies are starting to catch up with her, and the slow unraveling of Daemon's marriage while living in Pentos.
                      - generic [ref=e412]:
                        - generic [ref=e413]: 2022-09-25
                        - generic [ref=e414]: 7 min
                        - generic [ref=e415]:
                          - img [ref=e416]
                          - generic [ref=e418]: "4.0"
                    - button [ref=e420]:
                      - img [ref=e421]
                  - generic [ref=e423] [cursor=pointer]:
                    - img "Inside S1 E7" [ref=e425]
                    - generic [ref=e426]:
                      - generic [ref=e427]:
                        - generic [ref=e428]: S0E17
                        - heading "Inside S1 E7" [level=3] [ref=e429]
                      - paragraph [ref=e430]: Leo Ashton (Aemond Targaryen), Emma D'Arcy (Rhaenyra Targaryen), John MacMillian (Laenor Velaryon) and more speak to the fallout that occurs after Aemond loses his eye during a fight, and how Laena's death causes Laenor to reflect on his own life.
                      - generic [ref=e431]:
                        - generic [ref=e432]: 2022-10-02
                        - generic [ref=e433]: 7 min
                        - generic [ref=e434]:
                          - img [ref=e435]
                          - generic [ref=e437]: "6.0"
                    - button [ref=e439]:
                      - img [ref=e440]
                  - generic [ref=e442] [cursor=pointer]:
                    - img "Inside S1 E8" [ref=e444]
                    - generic [ref=e445]:
                      - generic [ref=e446]:
                        - generic [ref=e447]: S0E18
                        - heading "Inside S1 E8" [level=3] [ref=e448]
                      - paragraph [ref=e449]: Director Geeta Vasant Patel, Paddy Considine (Viserys Targaryen), Wil Johnson (Vaemond Velaryon) and more dive into the motivation behind Vaemond challenging Luke's claim to Driftmark, and Viserys' attempt at bringing his family together.
                      - generic [ref=e450]:
                        - generic [ref=e451]: 2022-10-09
                        - generic [ref=e452]: 9 min
                        - generic [ref=e453]:
                          - img [ref=e454]
                          - generic [ref=e456]: "6.0"
                    - button [ref=e458]:
                      - img [ref=e459]
                  - generic [ref=e461] [cursor=pointer]:
                    - img "Inside S1 E9" [ref=e463]
                    - generic [ref=e464]:
                      - generic [ref=e465]:
                        - generic [ref=e466]: S0E19
                        - heading "Inside S1 E9" [level=3] [ref=e467]
                      - paragraph [ref=e468]: Showrunners Ryan Condal and Miguel Sapochnik, director Clare Kilner, Tom Glynn-Carney (Aegon Targaryen) and more unpack the moment when Alicent finally sees her father for who he is, Aemond's underlying resentment for Aegon, and Rhaenys' escape from King's Landing.
                      - generic [ref=e469]:
                        - generic [ref=e470]: 2022-10-16
                        - generic [ref=e471]: 6 min
                        - generic [ref=e472]:
                          - img [ref=e473]
                          - generic [ref=e475]: "6.0"
                    - button [ref=e477]:
                      - img [ref=e478]
                  - generic [ref=e480] [cursor=pointer]:
                    - img "Inside S1 E10" [ref=e482]
                    - generic [ref=e483]:
                      - generic [ref=e484]:
                        - generic [ref=e485]: S0E20
                        - heading "Inside S1 E10" [level=3] [ref=e486]
                      - paragraph [ref=e487]: Showrunner Miguel Sapochnik, director Greg Yaitanes, Emma D'Arcy (Rhaenyra Targaryen), Eve Best (Rhaenys Targaryen), and more explore the emotional complexity Rhaenyra endures after hearing the Greens usurped the throne, Rhaenys' change of heart, and the implications of Aemond and Luke's encounter.
                      - generic [ref=e488]:
                        - generic [ref=e489]: 2022-10-23
                        - generic [ref=e490]: 10 min
                        - generic [ref=e491]:
                          - img [ref=e492]
                          - generic [ref=e494]: "4.0"
                    - button [ref=e496]:
                      - img [ref=e497]
                  - generic [ref=e499] [cursor=pointer]:
                    - img "Welcome To Westeros" [ref=e501]
                    - generic [ref=e502]:
                      - generic [ref=e503]:
                        - generic [ref=e504]: S0E21
                        - heading "Welcome To Westeros" [level=3] [ref=e505]
                      - paragraph [ref=e506]: Featurette from House of the Dragon season 1.
                      - generic [ref=e507]:
                        - generic [ref=e508]: 2022-12-20
                        - generic [ref=e509]: 6 min
                        - generic [ref=e510]:
                          - img [ref=e511]
                          - generic [ref=e513]: "4.0"
                    - button [ref=e515]:
                      - img [ref=e516]
                  - generic [ref=e518] [cursor=pointer]:
                    - img "A New Reign" [ref=e520]
                    - generic [ref=e521]:
                      - generic [ref=e522]:
                        - generic [ref=e523]: S0E22
                        - heading "A New Reign" [level=3] [ref=e524]
                      - paragraph [ref=e525]: George R. R. Martin, Ryan Condal and Miguel Sapochnik offer a look at what's to come in Game of Thrones prequel House of the Dragon.
                      - generic [ref=e526]:
                        - generic [ref=e527]: 2022-12-20
                        - generic [ref=e528]: 3 min
                        - generic [ref=e529]:
                          - img [ref=e530]
                          - generic [ref=e532]: "4.0"
                    - button [ref=e534]:
                      - img [ref=e535]
                  - generic [ref=e537] [cursor=pointer]:
                    - img "Returning to Westeros" [ref=e539]
                    - generic [ref=e540]:
                      - generic [ref=e541]:
                        - generic [ref=e542]: S0E23
                        - heading "Returning to Westeros" [level=3] [ref=e543]
                      - paragraph [ref=e544]: The cast and crew share the thrills and challenges of entering the world Game of Thrones, and approaching it when the Targaryens at the height of their power.
                      - generic [ref=e545]:
                        - generic [ref=e546]: 2022-12-20
                        - generic [ref=e547]: 5 min
                        - generic [ref=e548]:
                          - img [ref=e549]
                          - generic [ref=e551]: "4.0"
                    - button [ref=e553]:
                      - img [ref=e554]
                  - generic [ref=e556] [cursor=pointer]:
                    - 'img "Before the Dance: An Illustrated History with George R. R. Martin" [ref=e558]'
                    - generic [ref=e559]:
                      - generic [ref=e560]:
                        - generic [ref=e561]: S0E24
                        - 'heading "Before the Dance: An Illustrated History with George R. R. Martin" [level=3] [ref=e562]'
                      - paragraph [ref=e563]: Creator and executive producer George R. R. Martin sets up the Game of Thrones prequel, sharing key details on major Houses and relationships featured in the series, alongside illustrations from Fire & Blood artist Doug Wheatley.
                      - generic [ref=e564]:
                        - generic [ref=e565]: 2022-12-20
                        - generic [ref=e566]: 6 min
                        - generic [ref=e567]:
                          - img [ref=e568]
                          - generic [ref=e570]: "4.0"
                    - button [ref=e572]:
                      - img [ref=e573]
                  - generic [ref=e575] [cursor=pointer]:
                    - img "Height of an Empire" [ref=e577]
                    - generic [ref=e578]:
                      - generic [ref=e579]:
                        - generic [ref=e580]: S0E25
                        - heading "Height of an Empire" [level=3] [ref=e581]
                      - paragraph [ref=e582]: Paddy Considine, Matt Smith, and more cast and crew discuss how House of the Dragon takes place during the height of the Targaryen reign over Westeros, and how that influenced the story and design of the series.
                      - generic [ref=e583]:
                        - generic [ref=e584]: 2022-12-20
                        - generic [ref=e585]: 4 min
                        - generic [ref=e586]:
                          - img [ref=e587]
                          - generic [ref=e589]: "4.0"
                    - button [ref=e591]:
                      - img [ref=e592]
                  - generic [ref=e594] [cursor=pointer]:
                    - img "Noble Houses" [ref=e596]
                    - generic [ref=e597]:
                      - generic [ref=e598]:
                        - generic [ref=e599]: S0E26
                        - heading "Noble Houses" [level=3] [ref=e600]
                      - paragraph [ref=e601]: "Rhys Ifans, Steve Toussaint, and more cast and crew break down the historical significance and role of the two prominent houses in the series: the Velaryons and the Hightowers."
                      - generic [ref=e602]:
                        - generic [ref=e603]: 2022-12-20
                        - generic [ref=e604]: 4 min
                        - generic [ref=e605]:
                          - img [ref=e606]
                          - generic [ref=e608]: "4.0"
                    - button [ref=e610]:
                      - img [ref=e611]
                  - generic [ref=e613] [cursor=pointer]:
                    - img "Familiar Places" [ref=e615]
                    - generic [ref=e616]:
                      - generic [ref=e617]:
                        - generic [ref=e618]: S0E27
                        - heading "Familiar Places" [level=3] [ref=e619]
                      - paragraph [ref=e620]: The cast and crew describe the various changes made to recognizable sets like the Red Keep and the Throne to reflect the decadence of the time period and how they evolved over time.
                      - generic [ref=e621]:
                        - generic [ref=e622]: 2022-12-20
                        - generic [ref=e623]: 4 min
                        - generic [ref=e624]:
                          - img [ref=e625]
                          - generic [ref=e627]: "4.0"
                    - button [ref=e629]:
                      - img [ref=e630]
                  - generic [ref=e632] [cursor=pointer]:
                    - img "Return to the Seven Kingdoms" [ref=e634]
                    - generic [ref=e635]:
                      - generic [ref=e636]:
                        - generic [ref=e637]: S0E28
                        - heading "Return to the Seven Kingdoms" [level=3] [ref=e638]
                      - paragraph
                      - generic [ref=e639]:
                        - generic [ref=e640]: 2022-12-20
                        - generic [ref=e641]: 26 min
                        - generic [ref=e642]:
                          - img [ref=e643]
                          - generic [ref=e645]: "4.0"
                    - button [ref=e647]:
                      - img [ref=e648]
                  - generic [ref=e650] [cursor=pointer]:
                    - 'img "Introducing the Characters: Rhaenyra & Alicent" [ref=e652]'
                    - generic [ref=e653]:
                      - generic [ref=e654]:
                        - generic [ref=e655]: S0E29
                        - 'heading "Introducing the Characters: Rhaenyra & Alicent" [level=3] [ref=e656]'
                      - paragraph [ref=e657]: Olivia Cooke and Emma D'Arcy discuss their roles in the Game of Thrones prequel series House of the Dragon.
                      - generic [ref=e658]:
                        - generic [ref=e659]: 2022-12-20
                        - generic [ref=e660]: 2 min
                        - generic [ref=e661]:
                          - img [ref=e662]
                          - generic [ref=e664]: "4.0"
                    - button [ref=e666]:
                      - img [ref=e667]
                  - generic [ref=e669] [cursor=pointer]:
                    - 'img "Introducing the Characters: Corlys Velaryon" [ref=e671]'
                    - generic [ref=e672]:
                      - generic [ref=e673]:
                        - generic [ref=e674]: S0E30
                        - 'heading "Introducing the Characters: Corlys Velaryon" [level=3] [ref=e675]'
                      - paragraph [ref=e676]: Steve Toussaint discusses playing the role of Lord Corlys Velaryon, "The Sea Snake", in House of the Dragon.
                      - generic [ref=e677]:
                        - generic [ref=e678]: 2022-12-20
                        - generic [ref=e679]: 2 min
                        - generic [ref=e680]:
                          - img [ref=e681]
                          - generic [ref=e683]: "4.0"
                    - button [ref=e685]:
                      - img [ref=e686]
                  - generic [ref=e688] [cursor=pointer]:
                    - 'img "Introducing the Characters: Rhaenys Targaryen" [ref=e690]'
                    - generic [ref=e691]:
                      - generic [ref=e692]:
                        - generic [ref=e693]: S0E31
                        - 'heading "Introducing the Characters: Rhaenys Targaryen" [level=3] [ref=e694]'
                      - paragraph [ref=e695]: Eve Best discusses playing the role of Princess Rhaenys Targaryen in House of the Dragon.
                      - generic [ref=e696]:
                        - generic [ref=e697]: 2022-12-20
                        - generic [ref=e698]: 1 min
                        - generic [ref=e699]:
                          - img [ref=e700]
                          - generic [ref=e702]: "4.0"
                    - button [ref=e704]:
                      - img [ref=e705]
                  - generic [ref=e707] [cursor=pointer]:
                    - 'img "Introducing the Characters: Rhaenyra Targaryen" [ref=e709]'
                    - generic [ref=e710]:
                      - generic [ref=e711]:
                        - generic [ref=e712]: S0E32
                        - 'heading "Introducing the Characters: Rhaenyra Targaryen" [level=3] [ref=e713]'
                      - paragraph [ref=e714]: Emma D'Arcy discusses playing the role of Princess Rhaenyra Targaryen in House of the Dragon.
                      - generic [ref=e715]:
                        - generic [ref=e716]: 2022-12-20
                        - generic [ref=e717]: 1 min
                        - generic [ref=e718]:
                          - img [ref=e719]
                          - generic [ref=e721]: "4.0"
                    - button [ref=e723]:
                      - img [ref=e724]
                  - generic [ref=e726] [cursor=pointer]:
                    - 'img "Introducing the Characters: Viserys Targaryen" [ref=e728]'
                    - generic [ref=e729]:
                      - generic [ref=e730]:
                        - generic [ref=e731]: S0E33
                        - 'heading "Introducing the Characters: Viserys Targaryen" [level=3] [ref=e732]'
                      - paragraph [ref=e733]: Paddy Considine discusses playing the role of King Viserys Targaryen in House of the Dragon.
                      - generic [ref=e734]:
                        - generic [ref=e735]: 2022-12-20
                        - generic [ref=e736]: 1 min
                        - generic [ref=e737]:
                          - img [ref=e738]
                          - generic [ref=e740]: "4.0"
                    - button [ref=e742]:
                      - img [ref=e743]
                  - generic [ref=e745] [cursor=pointer]:
                    - 'img "Introducing the Characters: Daemon Targaryen" [ref=e747]'
                    - generic [ref=e748]:
                      - generic [ref=e749]:
                        - generic [ref=e750]: S0E34
                        - 'heading "Introducing the Characters: Daemon Targaryen" [level=3] [ref=e751]'
                      - paragraph [ref=e752]: Matt Smith discusses playing the role of Prince Daemon Targaryen in House of the Dragon.
                      - generic [ref=e753]:
                        - generic [ref=e754]: 2022-12-20
                        - generic [ref=e755]: 1 min
                        - generic [ref=e756]:
                          - img [ref=e757]
                          - generic [ref=e759]: "4.0"
                    - button [ref=e761]:
                      - img [ref=e762]
                  - generic [ref=e764] [cursor=pointer]:
                    - 'img "Introducing the Characters: Otto Hightower" [ref=e766]'
                    - generic [ref=e767]:
                      - generic [ref=e768]:
                        - generic [ref=e769]: S0E35
                        - 'heading "Introducing the Characters: Otto Hightower" [level=3] [ref=e770]'
                      - paragraph [ref=e771]: Rhys Ifans discusses playing the role of Otto Hightower in House of the Dragon.
                      - generic [ref=e772]:
                        - generic [ref=e773]: 2022-12-20
                        - generic [ref=e774]: 2 min
                        - generic [ref=e775]:
                          - img [ref=e776]
                          - generic [ref=e778]: "4.0"
                    - button [ref=e780]:
                      - img [ref=e781]
                  - generic [ref=e783] [cursor=pointer]:
                    - 'img "Introducing the Characters: Mysaria" [ref=e785]'
                    - generic [ref=e786]:
                      - generic [ref=e787]:
                        - generic [ref=e788]: S0E36
                        - 'heading "Introducing the Characters: Mysaria" [level=3] [ref=e789]'
                      - paragraph [ref=e790]: Sonoya Mizuno discusses playing the role of Mysaria in House of the Dragon.
                      - generic [ref=e791]:
                        - generic [ref=e792]: 2022-12-20
                        - generic [ref=e793]: 2 min
                        - generic [ref=e794]:
                          - img [ref=e795]
                          - generic [ref=e797]: "4.0"
                    - button [ref=e799]:
                      - img [ref=e800]
                  - generic [ref=e802] [cursor=pointer]:
                    - 'img "Introducing the Characters: Criston Cole" [ref=e804]'
                    - generic [ref=e805]:
                      - generic [ref=e806]:
                        - generic [ref=e807]: S0E37
                        - 'heading "Introducing the Characters: Criston Cole" [level=3] [ref=e808]'
                      - paragraph [ref=e809]: Fabien Frankel discusses playing the role of Ser Criston Cole in House of the Dragon.
                      - generic [ref=e810]:
                        - generic [ref=e811]: 2022-12-20
                        - generic [ref=e812]: 2 min
                        - generic [ref=e813]:
                          - img [ref=e814]
                          - generic [ref=e816]: "4.0"
                    - button [ref=e818]:
                      - img [ref=e819]
                  - generic [ref=e821] [cursor=pointer]:
                    - img "The House That Dragons Built S2 E1" [ref=e823]
                    - generic [ref=e824]:
                      - generic [ref=e825]:
                        - generic [ref=e826]: S0E38
                        - heading "The House That Dragons Built S2 E1" [level=3] [ref=e827]
                      - paragraph [ref=e828]: Showrunner Ryan Condal, Director Alan Taylor, Production Designer Jim Clay, and more reflect on the making of the making of season 2 episode 1, "A Son for a Son".
                      - generic [ref=e829]:
                        - generic [ref=e830]: 2024-06-16
                        - generic [ref=e831]: 20 min
                        - generic [ref=e832]:
                          - img [ref=e833]
                          - generic [ref=e835]: "0.0"
                    - button [ref=e837]:
                      - img [ref=e838]
                  - generic [ref=e840] [cursor=pointer]:
                    - img "The House That Dragons Built S2 E2" [ref=e842]
                    - generic [ref=e843]:
                      - generic [ref=e844]:
                        - generic [ref=e845]: S0E39
                        - heading "The House That Dragons Built S2 E2" [level=3] [ref=e846]
                      - paragraph [ref=e847]: Showrunner Ryan Condal, director Clare Kilner, actress Olivia Cooke (Queen Alicent Hightower), and more take us behind the scenes of the chaotic opening of episode 2, and the funeral scene where Queen Alicent Hightower and Queen Helaena Targaryen's (Phia Saban) emotions are on full display.
                      - generic [ref=e848]:
                        - generic [ref=e849]: 2024-06-23
                        - generic [ref=e850]: 21 min
                        - generic [ref=e851]:
                          - img [ref=e852]
                          - generic [ref=e854]: "0.0"
                    - button [ref=e856]:
                      - img [ref=e857]
                  - generic [ref=e859] [cursor=pointer]:
                    - img "The House That Dragons Built S2 E3" [ref=e861]
                    - generic [ref=e862]:
                      - generic [ref=e863]:
                        - generic [ref=e864]: S0E40
                        - heading "The House That Dragons Built S2 E3" [level=3] [ref=e865]
                      - paragraph [ref=e866]: Director Geeta Vasant Patel, production designer Jim Clay, actor Matt Smith (Daemon Targaryen), and more reflect on the scale of the Harrenhal castle set, the introduction and design of Baela’s (Bethany Antonia) dragon, Moondancer, and the creation of Aegon's (Tom Glynn-Carney) ancient Valyrian armor.
                      - generic [ref=e867]:
                        - generic [ref=e868]: 2024-06-30
                        - generic [ref=e869]: 20 min
                        - generic [ref=e870]:
                          - img [ref=e871]
                          - generic [ref=e873]: "0.0"
                    - button [ref=e875]:
                      - img [ref=e876]
                  - generic [ref=e878] [cursor=pointer]:
                    - img "The House That Dragons Built S2 E4" [ref=e880]
                    - generic [ref=e881]:
                      - generic [ref=e882]:
                        - generic [ref=e883]: S0E41
                        - heading "The House That Dragons Built S2 E4" [level=3] [ref=e884]
                      - paragraph [ref=e885]: Showrunner Ryan Condal, director Alan Taylor, and actress Eve Best (Rhaenys Targaryen) reflect on the massive scale of the Battle at Rook's Rest and the devastating consequences of dragons joining the battle.
                      - generic [ref=e886]:
                        - generic [ref=e887]: 2024-07-07
                        - generic [ref=e888]: 20 min
                        - generic [ref=e889]:
                          - img [ref=e890]
                          - generic [ref=e892]: "0.0"
                    - button [ref=e894]:
                      - img [ref=e895]
                  - generic [ref=e897] [cursor=pointer]:
                    - img "The House That Dragons Built S2 E5" [ref=e899]
                    - generic [ref=e900]:
                      - generic [ref=e901]:
                        - generic [ref=e902]: S0E42
                        - heading "The House That Dragons Built S2 E5" [level=3] [ref=e903]
                      - paragraph [ref=e904]: Director Clare Kilner, director of photography Alejandro Martínez, actor Steve Toussaint (Lord Corlys), and more discuss the aftermath of the battle at Rook's Rest and the immersive filming locations of episode 5.
                      - generic [ref=e905]:
                        - generic [ref=e906]: 2024-07-14
                        - generic [ref=e907]: 20 min
                        - generic [ref=e908]:
                          - img [ref=e909]
                          - generic [ref=e911]: "0.0"
                    - button [ref=e913]:
                      - img [ref=e914]
                  - generic [ref=e916] [cursor=pointer]:
                    - img "The House That Dragons Built S2 E6" [ref=e918]
                    - generic [ref=e919]:
                      - generic [ref=e920]:
                        - generic [ref=e921]: S0E43
                        - heading "The House That Dragons Built S2 E6" [level=3] [ref=e922]
                      - paragraph [ref=e923]: Showrunner Ryan Condal, stunt coordinator Rowley Irlam, special effects supervisor Michael Dawson, and more discuss the complex stunt work and innovative pyrotechnics used to create Ser Steffon's encounter with Seasmoke.
                      - generic [ref=e924]:
                        - generic [ref=e925]: 2024-07-21
                        - generic [ref=e926]: 20 min
                        - generic [ref=e927]:
                          - img [ref=e928]
                          - generic [ref=e930]: "0.0"
                    - button [ref=e932]:
                      - img [ref=e933]
                  - generic [ref=e935] [cursor=pointer]:
                    - img "The House That Dragons Built S2 E7" [ref=e937]
                    - generic [ref=e938]:
                      - generic [ref=e939]:
                        - generic [ref=e940]: S0E44
                        - heading "The House That Dragons Built S2 E7" [level=3] [ref=e941]
                      - paragraph [ref=e942]: Director Loni Peristere, production designer Jim Clay, VFX producer Thomas Horton, and more reflect on the creation of the Godswoods at Harrenhal, the intense stunt work behind Vermithor's attack, and the cutting-edge visual effects behind riding dragons.
                      - generic [ref=e943]:
                        - generic [ref=e944]: 2024-07-28
                        - generic [ref=e945]: 20 min
                        - generic [ref=e946]:
                          - img [ref=e947]
                          - generic [ref=e949]: "0.0"
                    - button [ref=e951]:
                      - img [ref=e952]
                  - generic [ref=e954] [cursor=pointer]:
                    - img "The House That Dragons Built S2 E8" [ref=e956]
                    - generic [ref=e957]:
                      - generic [ref=e958]:
                        - generic [ref=e959]: S0E45
                        - heading "The House That Dragons Built S2 E8" [level=3] [ref=e960]
                      - paragraph [ref=e961]: Showrunner Ryan Condal, director Geeta Vasant Patel, and more dissect the revelations of Daemon's (Matt Smith) final vision at Harrenhal, the visual effects used to grow Rhaenyra's (Emma D'Arcy) army, and the innovative special effects used to create the impactful montage at the end of season 2.
                      - generic [ref=e962]:
                        - generic [ref=e963]: 2024-08-04
                        - generic [ref=e964]: 20 min
                        - generic [ref=e965]:
                          - img [ref=e966]
                          - generic [ref=e968]: "0.0"
                    - button [ref=e970]:
                      - img [ref=e971]
                  - generic [ref=e973] [cursor=pointer]:
                    - img "Inside S2 E1" [ref=e975]
                    - generic [ref=e976]:
                      - generic [ref=e977]:
                        - generic [ref=e978]: S0E46
                        - heading "Inside S2 E1" [level=3] [ref=e979]
                      - paragraph [ref=e980]: Showrunner Ryan Condal, director Alan Taylor, and the cast discuss where the story picks up in the new season, finding Rhaenyra deep in grief, Aegon facing his first challenges as king, and Alicent seeking comfort in an unexpected companion.
                      - generic [ref=e981]:
                        - generic [ref=e982]: 2024-06-16
                        - generic [ref=e983]: 9 min
                        - generic [ref=e984]:
                          - img [ref=e985]
                          - generic [ref=e987]: "0.0"
                    - button [ref=e989]:
                      - img [ref=e990]
                  - generic [ref=e992] [cursor=pointer]:
                    - img "Inside S2 E2" [ref=e994]
                    - generic [ref=e995]:
                      - generic [ref=e996]:
                        - generic [ref=e997]: S0E47
                        - heading "Inside S2 E2" [level=3] [ref=e998]
                      - paragraph [ref=e999]: Showrunner Ryan Condal, writer Sarah Hess, and more discuss Aegon's (Tom Glynn-Carney) lost sense of stability, Rhaenyra (Emma D'Arcy) embracing her rage, and the plot set in motion by Ser Criston (Fabien Frankel), culminating in the epic showdown between Ser Erryk and Ser Arryk.
                      - generic [ref=e1000]:
                        - generic [ref=e1001]: 2024-06-23
                        - generic [ref=e1002]: 8 min
                        - generic [ref=e1003]:
                          - img [ref=e1004]
                          - generic [ref=e1006]: "0.0"
                    - button [ref=e1008]:
                      - img [ref=e1009]
                  - generic [ref=e1011] [cursor=pointer]:
                    - img "Inside S2 E3" [ref=e1013]
                    - generic [ref=e1014]:
                      - generic [ref=e1015]:
                        - generic [ref=e1016]: S0E48
                        - heading "Inside S2 E3" [level=3] [ref=e1017]
                      - paragraph [ref=e1018]: Showrunner Ryan Condal, director Geeta Vasant Patel, and more dissect Ser Criston's (Fabien Frankel) anxiety-inducing first day as Hand of the King, Daemon’s (Matt Smith) supernatural hallucinations about his past, Rhaena's (Phone Campbell) resentment towards not having an important role in the war, and the long awaited face-to-face conversation between Rhaenyra (Emma D'Arcy) and Alicent (Olivia Cooke).
                      - generic [ref=e1019]:
                        - generic [ref=e1020]: 2024-06-30
                        - generic [ref=e1021]: 9 min
                        - generic [ref=e1022]:
                          - img [ref=e1023]
                          - generic [ref=e1025]: "0.0"
                    - button [ref=e1027]:
                      - img [ref=e1028]
                  - generic [ref=e1030] [cursor=pointer]:
                    - img "Inside S2 E4" [ref=e1032]
                    - generic [ref=e1033]:
                      - generic [ref=e1034]:
                        - generic [ref=e1035]: S0E49
                        - heading "Inside S2 E4" [level=3] [ref=e1036]
                      - paragraph [ref=e1037]: Showrunner Ryan Condal, director Alan Taylor, actor Ewan Mitchell (Aemond Targaryen), and more discuss the escalating tensions between characters that comes to a head during the epic Battle at Rook's Rest.
                      - generic [ref=e1038]:
                        - generic [ref=e1039]: 2024-07-07
                        - generic [ref=e1040]: 9 min
                        - generic [ref=e1041]:
                          - img [ref=e1042]
                          - generic [ref=e1044]: "0.0"
                    - button [ref=e1046]:
                      - img [ref=e1047]
                  - generic [ref=e1049] [cursor=pointer]:
                    - img "Inside S2 E5" [ref=e1051]
                    - generic [ref=e1052]:
                      - generic [ref=e1053]:
                        - generic [ref=e1054]: S0E50
                        - heading "Inside S2 E5" [level=3] [ref=e1055]
                      - paragraph [ref=e1056]: Showrunner Ryan Condal, director Clare Kilner, and more reflect on the aftermath of the Battle at Rook's Rest and the shifting power dynamics on both sides of the war.
                      - generic [ref=e1057]:
                        - generic [ref=e1058]: 2024-07-14
                        - generic [ref=e1059]: 7 min
                        - generic [ref=e1060]:
                          - img [ref=e1061]
                          - generic [ref=e1063]: "0.0"
                    - button [ref=e1065]:
                      - img [ref=e1066]
                  - generic [ref=e1068] [cursor=pointer]:
                    - img "Inside S2 E6" [ref=e1070]
                    - generic [ref=e1071]:
                      - generic [ref=e1072]:
                        - generic [ref=e1073]: S0E51
                        - heading "Inside S2 E6" [level=3] [ref=e1074]
                      - paragraph [ref=e1075]: Showrunner Ryan Condal, director Andrij Parekh, and more reflect on Aemond's (Ewan Mitchell) newfound power, the rising tensions amongst the people of King's Landing, and the changing dynamics between Mysaria (Sonoya Mizuno) and Rhaenyra (Emma D’Arcy).
                      - generic [ref=e1076]:
                        - generic [ref=e1077]: 2024-07-21
                        - generic [ref=e1078]: 8 min
                        - generic [ref=e1079]:
                          - img [ref=e1080]
                          - generic [ref=e1082]: "0.0"
                    - button [ref=e1084]:
                      - img [ref=e1085]
                  - generic [ref=e1087] [cursor=pointer]:
                    - img "Inside S2 E7" [ref=e1089]
                    - generic [ref=e1090]:
                      - generic [ref=e1091]:
                        - generic [ref=e1092]: S0E52
                        - heading "Inside S2 E7" [level=3] [ref=e1093]
                      - paragraph [ref=e1094]: Showrunner Ryan Condal, actor Emma D'Arcy (Rhaenyra), actor Clinton Liberty (Addam), and more discuss Rhaenyra and Syrax's standoff with Addam and Seasmoke, and Aemond's (Ewan Mitchell) shock when he discovers Rhaenyra's army of dragons.
                      - generic [ref=e1095]:
                        - generic [ref=e1096]: 2024-07-28
                        - generic [ref=e1097]: 8 min
                        - generic [ref=e1098]:
                          - img [ref=e1099]
                          - generic [ref=e1101]: "0.0"
                    - button [ref=e1103]:
                      - img [ref=e1104]
                  - generic [ref=e1106] [cursor=pointer]:
                    - img "Inside S2 E8" [ref=e1108]
                    - generic [ref=e1109]:
                      - generic [ref=e1110]:
                        - generic [ref=e1111]: S0E53
                        - heading "Inside S2 E8" [level=3] [ref=e1112]
                      - paragraph [ref=e1113]: Showrunner Ryan Condal, writer Sara Hess, director Geeta Vasant Patel, and more reflect on Aemond's (Ewan Mitchell) fear of losing power as his secret comes to light, Daemon's (Matt Smith) evolution after his time at Harrenhal, and the difficult decisions made by Alicent (Olivia Cooke) and Aegon (Tom Glynn-Carney) that set the tone for Season 3.
                      - generic [ref=e1114]:
                        - generic [ref=e1115]: 2024-08-04
                        - generic [ref=e1116]: 13 min
                        - generic [ref=e1117]:
                          - img [ref=e1118]
                          - generic [ref=e1120]: "0.0"
                    - button [ref=e1122]:
                      - img [ref=e1123]
                  - generic [ref=e1125] [cursor=pointer]:
                    - img "House Who? House Stark" [ref=e1127]
                    - generic [ref=e1128]:
                      - generic [ref=e1129]:
                        - generic [ref=e1130]: S0E54
                        - heading "House Who? House Stark" [level=3] [ref=e1131]
                      - paragraph [ref=e1132]: "\"The North must stand ready.\" Showrunner Ryan Condal discusses Season 2's return to the Wall and the importance of the Starks in the escalating conflict."
                      - generic [ref=e1133]:
                        - generic [ref=e1134]: 2024-11-21
                        - generic [ref=e1135]: 2 min
                        - generic [ref=e1136]:
                          - img [ref=e1137]
                          - generic [ref=e1139]: "0.0"
                    - button [ref=e1141]:
                      - img [ref=e1142]
                  - generic [ref=e1144] [cursor=pointer]:
                    - img "House Who? Bracken & Blackwood" [ref=e1146]
                    - generic [ref=e1147]:
                      - generic [ref=e1148]:
                        - generic [ref=e1149]: S0E55
                        - heading "House Who? Bracken & Blackwood" [level=3] [ref=e1150]
                      - paragraph [ref=e1151]: Showrunner Ryan Condal and director Geeta Vasant Patel discuss the centuries-long feud between the Blackwoods and the Brackens that comes to a head at the Battle of the Burning Mill.
                      - generic [ref=e1152]:
                        - generic [ref=e1153]: 2024-11-21
                        - generic [ref=e1154]: 2 min
                        - generic [ref=e1155]:
                          - img [ref=e1156]
                          - generic [ref=e1158]: "0.0"
                    - button [ref=e1160]:
                      - img [ref=e1161]
                  - generic [ref=e1163] [cursor=pointer]:
                    - img "House Who? Tully & Frey" [ref=e1165]
                    - generic [ref=e1166]:
                      - generic [ref=e1167]:
                        - generic [ref=e1168]: S0E56
                        - heading "House Who? Tully & Frey" [level=3] [ref=e1169]
                      - paragraph [ref=e1170]: Showrunner Ryan Condal, writer David Hancock, actor Matt Smith (Daemon Targaryen), and more explain the important role House Tully and House Frey play in helping the Black Council gain support in the Riverlands.
                      - generic [ref=e1171]:
                        - generic [ref=e1172]: 2024-11-21
                        - generic [ref=e1173]: 2 min
                        - generic [ref=e1174]:
                          - img [ref=e1175]
                          - generic [ref=e1177]: "0.0"
                    - button [ref=e1179]:
                      - img [ref=e1180]
                  - generic [ref=e1182] [cursor=pointer]:
                    - img "Guess that Line - Eve & Steve" [ref=e1184]
                    - generic [ref=e1185]:
                      - generic [ref=e1186]:
                        - generic [ref=e1187]: S0E57
                        - heading "Guess that Line - Eve & Steve" [level=3] [ref=e1188]
                      - paragraph [ref=e1189]: Eve and Steve attempt to guess who said specific lines in Season 1.
                      - generic [ref=e1190]:
                        - generic [ref=e1191]: 2024-11-21
                        - generic [ref=e1192]: 4 min
                        - generic [ref=e1193]:
                          - img [ref=e1194]
                          - generic [ref=e1196]: "0.0"
                    - button [ref=e1198]:
                      - img [ref=e1199]
                  - generic [ref=e1201] [cursor=pointer]:
                    - img "Eve Tribute Piece" [ref=e1203]
                    - generic [ref=e1204]:
                      - generic [ref=e1205]:
                        - generic [ref=e1206]: S0E58
                        - heading "Eve Tribute Piece" [level=3] [ref=e1207]
                      - paragraph [ref=e1208]: Showrunner Ryan Condal, actress Eve Best (Rhaenys Targaryen), and actor Steve Toussaint (Lord Corlys) pay tribute to Princess Rhaenys, after her death during the Battle at Rook’s Rest.
                      - generic [ref=e1209]:
                        - generic [ref=e1210]: 2024-11-21
                        - generic [ref=e1211]: 2 min
                        - generic [ref=e1212]:
                          - img [ref=e1213]
                          - generic [ref=e1215]: "0.0"
                    - button [ref=e1217]:
                      - img [ref=e1218]
                  - generic [ref=e1220] [cursor=pointer]:
                    - img "Targaryen Family Tree" [ref=e1222]
                    - generic [ref=e1223]:
                      - generic [ref=e1224]:
                        - generic [ref=e1225]: S0E59
                        - heading "Targaryen Family Tree" [level=3] [ref=e1226]
                      - paragraph [ref=e1227]: Westerosi family ties can be confusing. This primer will help viewers keep track of the complex and far-reaching branches of House Targaryen.
                      - generic [ref=e1228]:
                        - generic [ref=e1229]: 2024-11-21
                        - generic [ref=e1230]: 6 min
                        - generic [ref=e1231]:
                          - img [ref=e1232]
                          - generic [ref=e1234]: "0.0"
                    - button [ref=e1236]:
                      - img [ref=e1237]
                  - generic [ref=e1239] [cursor=pointer]:
                    - img "The Curse of Harrenhal" [ref=e1241]
                    - generic [ref=e1242]:
                      - generic [ref=e1243]:
                        - generic [ref=e1244]: S0E60
                        - heading "The Curse of Harrenhal" [level=3] [ref=e1245]
                      - paragraph [ref=e1246]: Showrunner Ryan Condal, director Geeta Vasant Patel, actor Matt Smith (Daemon Targaryen), and more discuss the mysterious, seemingly cursed Harrenhal castle, as well as Daemon's personal journey of discovery and unraveling.
                      - generic [ref=e1247]:
                        - generic [ref=e1248]: 2024-11-21
                        - generic [ref=e1249]: 8 min
                        - generic [ref=e1250]:
                          - img [ref=e1251]
                          - generic [ref=e1253]: "0.0"
                    - button [ref=e1255]:
                      - img [ref=e1256]
                  - generic [ref=e1258] [cursor=pointer]:
                    - img "Divided Kingdoms" [ref=e1260]
                    - generic [ref=e1261]:
                      - generic [ref=e1262]:
                        - generic [ref=e1263]: S0E61
                        - heading "Divided Kingdoms" [level=3] [ref=e1264]
                      - paragraph [ref=e1265]: Join Co-Creator/Showrunner/Executive Producer Ryan Condal and the cast and crew as they provide an overview of Season 2 and a glimpse of the war to come.
                      - generic [ref=e1266]:
                        - generic [ref=e1267]: 2024-11-21
                        - generic [ref=e1268]: 9 min
                        - generic [ref=e1269]:
                          - img [ref=e1270]
                          - generic [ref=e1272]: "0.0"
                    - button [ref=e1274]:
                      - img [ref=e1275]
                  - generic [ref=e1277] [cursor=pointer]:
                    - img "Defend Your Council" [ref=e1279]
                    - generic [ref=e1280]:
                      - generic [ref=e1281]:
                        - generic [ref=e1282]: S0E62
                        - heading "Defend Your Council" [level=3] [ref=e1283]
                      - paragraph [ref=e1284]: The cast has pledged their loyalties. Now, it’s your turn.
                      - generic [ref=e1285]:
                        - generic [ref=e1286]: 2024-11-21
                        - generic [ref=e1287]: 2 min
                        - generic [ref=e1288]:
                          - img [ref=e1289]
                          - generic [ref=e1291]: "0.0"
                    - button [ref=e1293]:
                      - img [ref=e1294]
                  - generic [ref=e1296] [cursor=pointer]:
                    - img "Return to Winterfell" [ref=e1298]
                    - generic [ref=e1299]:
                      - generic [ref=e1300]:
                        - generic [ref=e1301]: S0E63
                        - heading "Return to Winterfell" [level=3] [ref=e1302]
                      - paragraph [ref=e1303]: Showrunner Ryan Condal, Production Designer Jim Clay, and more talk about the full circle moment of returning to Winterfell for House of The Dragon Season 2.
                      - generic [ref=e1304]:
                        - generic [ref=e1305]: 2024-11-21
                        - generic [ref=e1306]: 1 min
                        - generic [ref=e1307]:
                          - img [ref=e1308]
                          - generic [ref=e1310]: "0.0"
                    - button [ref=e1312]:
                      - img [ref=e1313]
                  - generic [ref=e1315] [cursor=pointer]:
                    - img "Return to the Realm" [ref=e1317]
                    - generic [ref=e1318]:
                      - generic [ref=e1319]:
                        - generic [ref=e1320]: S0E64
                        - heading "Return to the Realm" [level=3] [ref=e1321]
                      - paragraph [ref=e1322]: Showrunner Ryan Condal and the cast express their excitement to be back on set and what fans can expect in Season 2.
                      - generic [ref=e1323]:
                        - generic [ref=e1324]: 2024-11-21
                        - generic [ref=e1325]: 2 min
                        - generic [ref=e1326]:
                          - img [ref=e1327]
                          - generic [ref=e1329]: "0.0"
                    - button [ref=e1331]:
                      - img [ref=e1332]
                  - generic [ref=e1334] [cursor=pointer]:
                    - 'img "Fire Hot Takes: Team Green vs. Team Black" [ref=e1336]'
                    - generic [ref=e1337]:
                      - generic [ref=e1338]:
                        - generic [ref=e1339]: S0E65
                        - 'heading "Fire Hot Takes: Team Green vs. Team Black" [level=3] [ref=e1340]'
                      - paragraph [ref=e1341]: Team Green or Team Black? A choice must be made.
                      - generic [ref=e1342]:
                        - generic [ref=e1343]: 2024-11-21
                        - generic [ref=e1344]: 2 min
                        - generic [ref=e1345]:
                          - img [ref=e1346]
                          - generic [ref=e1348]: "0.0"
                    - button [ref=e1350]:
                      - img [ref=e1351]
                  - generic [ref=e1353] [cursor=pointer]:
                    - 'img "Fire Hot Takes: Case for Ruler" [ref=e1355]'
                    - generic [ref=e1356]:
                      - generic [ref=e1357]:
                        - generic [ref=e1358]: S0E66
                        - 'heading "Fire Hot Takes: Case for Ruler" [level=3] [ref=e1359]'
                      - paragraph [ref=e1360]: Who do you think would make a better king or queen? All Must Choose.
                      - generic [ref=e1361]:
                        - generic [ref=e1362]: 2024-11-21
                        - generic [ref=e1363]: 2 min
                        - generic [ref=e1364]:
                          - img [ref=e1365]
                          - generic [ref=e1367]: "0.0"
                    - button [ref=e1369]:
                      - img [ref=e1370]
                  - generic [ref=e1372] [cursor=pointer]:
                    - 'img "Fire Hot Takes: Daemon vs Aemond" [ref=e1374]'
                    - generic [ref=e1375]:
                      - generic [ref=e1376]:
                        - generic [ref=e1377]: S0E67
                        - 'heading "Fire Hot Takes: Daemon vs Aemond" [level=3] [ref=e1378]'
                      - paragraph [ref=e1379]: Team Black and Team Green weigh in on "who has less fucks to give" in House of the Dragon.
                      - generic [ref=e1380]:
                        - generic [ref=e1381]: 2024-11-21
                        - generic [ref=e1382]: 2 min
                        - generic [ref=e1383]:
                          - img [ref=e1384]
                          - generic [ref=e1386]: "0.0"
                    - button [ref=e1388]:
                      - img [ref=e1389]
                  - generic [ref=e1391] [cursor=pointer]:
                    - 'img "Character Spots: Rhaenyra" [ref=e1393]'
                    - generic [ref=e1394]:
                      - generic [ref=e1395]:
                        - generic [ref=e1396]: S0E68
                        - 'heading "Character Spots: Rhaenyra" [level=3] [ref=e1397]'
                      - paragraph [ref=e1398]: Emma D'Arcy discusses what's next for her character Rhaenyra Targaryen after the devastating events of the Season 1 finale.
                      - generic [ref=e1399]:
                        - generic [ref=e1400]: 2024-11-21
                        - generic [ref=e1401]: 1 min
                        - generic [ref=e1402]:
                          - img [ref=e1403]
                          - generic [ref=e1405]: "0.0"
                    - button [ref=e1407]:
                      - img [ref=e1408]
                  - generic [ref=e1410] [cursor=pointer]:
                    - 'img "Character Spots: Aegon" [ref=e1412]'
                    - generic [ref=e1413]:
                      - generic [ref=e1414]:
                        - generic [ref=e1415]: S0E69
                        - 'heading "Character Spots: Aegon" [level=3] [ref=e1416]'
                      - paragraph [ref=e1417]: Tom Glynn-Carney looks ahead at what Season 2 has in store for his character Aegon Targaryen.
                      - generic [ref=e1418]:
                        - generic [ref=e1419]: 2024-11-21
                        - generic [ref=e1420]: 1 min
                        - generic [ref=e1421]:
                          - img [ref=e1422]
                          - generic [ref=e1424]: "0.0"
                    - button [ref=e1426]:
                      - img [ref=e1427]
                  - generic [ref=e1429] [cursor=pointer]:
                    - 'img "Character Spots: Daemon" [ref=e1431]'
                    - generic [ref=e1432]:
                      - generic [ref=e1433]:
                        - generic [ref=e1434]: S0E70
                        - 'heading "Character Spots: Daemon" [level=3] [ref=e1435]'
                      - paragraph [ref=e1436]: Matt Smith discusses his character Prince Daemon Targaryen and what's to come in Season 2.
                      - generic [ref=e1437]:
                        - generic [ref=e1438]: 2024-11-21
                        - generic [ref=e1439]: 1 min
                        - generic [ref=e1440]:
                          - img [ref=e1441]
                          - generic [ref=e1443]: "0.0"
                    - button [ref=e1445]:
                      - img [ref=e1446]
                  - generic [ref=e1448] [cursor=pointer]:
                    - 'img "Character Spots: Corlys" [ref=e1450]'
                    - generic [ref=e1451]:
                      - generic [ref=e1452]:
                        - generic [ref=e1453]: S0E71
                        - 'heading "Character Spots: Corlys" [level=3] [ref=e1454]'
                      - paragraph [ref=e1455]: Steve Toussaint discusses playing the role of Lord Corlys Velaryon, "The Sea Snake", in House of the Dragon.
                      - generic [ref=e1456]:
                        - generic [ref=e1457]: 2024-11-21
                        - generic [ref=e1458]: 1 min
                        - generic [ref=e1459]:
                          - img [ref=e1460]
                          - generic [ref=e1462]: "0.0"
                    - button [ref=e1464]:
                      - img [ref=e1465]
                  - generic [ref=e1467] [cursor=pointer]:
                    - 'img "Character Spots: Alicent" [ref=e1469]'
                    - generic [ref=e1470]:
                      - generic [ref=e1471]:
                        - generic [ref=e1472]: S0E72
                        - 'heading "Character Spots: Alicent" [level=3] [ref=e1473]'
                      - paragraph [ref=e1474]: Olivia Cooke discusses what's to come for her character Queen Alicent Hightower in Season 2.
                      - generic [ref=e1475]:
                        - generic [ref=e1476]: 2024-11-21
                        - generic [ref=e1477]: 2 min
                        - generic [ref=e1478]:
                          - img [ref=e1479]
                          - generic [ref=e1481]: "0.0"
                    - button [ref=e1483]:
                      - img [ref=e1484]
                  - generic [ref=e1486] [cursor=pointer]:
                    - 'img "Character Spots: Aemond" [ref=e1488]'
                    - generic [ref=e1489]:
                      - generic [ref=e1490]:
                        - generic [ref=e1491]: S0E73
                        - 'heading "Character Spots: Aemond" [level=3] [ref=e1492]'
                      - paragraph [ref=e1493]: Ewan Mitchell discusses what's next for his character Prince Aemond Targaryen in the wake of the shocking events of the Season 1 finale.
                      - generic [ref=e1494]:
                        - generic [ref=e1495]: 2024-11-21
                        - generic [ref=e1496]: 2 min
                        - generic [ref=e1497]:
                          - img [ref=e1498]
                          - generic [ref=e1500]: "0.0"
                    - button [ref=e1502]:
                      - img [ref=e1503]
              - generic [ref=e1505]:
                - heading "Cast" [level=2] [ref=e1506]
                - generic [ref=e1507]:
                  - generic [ref=e1508]:
                    - img "Matt Smith" [ref=e1510]
                    - paragraph [ref=e1511]: Matt Smith
                    - paragraph [ref=e1512]: Prince Daemon Targaryen
                  - generic [ref=e1513]:
                    - img "Emma D'Arcy" [ref=e1515]
                    - paragraph [ref=e1516]: Emma D'Arcy
                    - paragraph [ref=e1517]: Princess Rhaenyra Targaryen
                  - generic [ref=e1518]:
                    - img "Olivia Cooke" [ref=e1520]
                    - paragraph [ref=e1521]: Olivia Cooke
                    - paragraph [ref=e1522]: Queen Alicent Hightower
                  - generic [ref=e1523]:
                    - img "Steve Toussaint" [ref=e1525]
                    - paragraph [ref=e1526]: Steve Toussaint
                    - paragraph [ref=e1527]: Lord Corlys 'The Sea Snake' Velaryon
                  - generic [ref=e1528]:
                    - img "Rhys Ifans" [ref=e1530]
                    - paragraph [ref=e1531]: Rhys Ifans
                    - paragraph [ref=e1532]: Ser Otto Hightower
                  - generic [ref=e1533]:
                    - img "Fabien Frankel" [ref=e1535]
                    - paragraph [ref=e1536]: Fabien Frankel
                    - paragraph [ref=e1537]: Ser Criston Cole
                  - generic [ref=e1538]:
                    - img "Ewan Mitchell" [ref=e1540]
                    - paragraph [ref=e1541]: Ewan Mitchell
                    - paragraph [ref=e1542]: Prince Aemond Targaryen
                  - generic [ref=e1543]:
                    - img "Tom Glynn-Carney" [ref=e1545]
                    - paragraph [ref=e1546]: Tom Glynn-Carney
                    - paragraph [ref=e1547]: King Aegon II Targaryen
                  - generic [ref=e1548]:
                    - img "Sonoya Mizuno" [ref=e1550]
                    - paragraph [ref=e1551]: Sonoya Mizuno
                    - paragraph [ref=e1552]: Mysaria 'The White Worm'
                  - generic [ref=e1553]:
                    - img "Harry Collett" [ref=e1555]
                    - paragraph [ref=e1556]: Harry Collett
                    - paragraph [ref=e1557]: Prince Jacaerys 'Jace' Velaryon
            - generic [ref=e1559]:
              - heading "Show Info" [level=3] [ref=e1560]
              - generic [ref=e1561]:
                - generic [ref=e1562]:
                  - paragraph [ref=e1563]: First Air Date
                  - paragraph [ref=e1564]: 2022-08-21
                - generic [ref=e1565]:
                  - paragraph [ref=e1566]: Last Air Date
                  - paragraph [ref=e1567]: 2026-06-21
                - generic [ref=e1568]:
                  - paragraph [ref=e1569]: Status
                  - paragraph [ref=e1570]: Returning Series
                - generic [ref=e1571]:
                  - paragraph [ref=e1572]: Type
                  - paragraph [ref=e1573]: Scripted
                - generic [ref=e1574]:
                  - paragraph [ref=e1575]: Seasons
                  - paragraph [ref=e1576]: "3"
                - generic [ref=e1577]:
                  - paragraph [ref=e1578]: Episodes
                  - paragraph [ref=e1579]: "26"
                - generic [ref=e1580]:
                  - paragraph [ref=e1581]: Created By
                  - paragraph [ref=e1582]: George R.R. Martin, Ryan J. Condal
                - generic [ref=e1583]:
                  - paragraph [ref=e1584]: Network
                  - paragraph [ref=e1585]: HBO
          - generic [ref=e1586]:
            - heading "Similar Shows" [level=2] [ref=e1587]
            - generic [ref=e1588]:
              - link "Titus Titus 6.9" [ref=e1589] [cursor=pointer]:
                - /url: /tv/89
                - img "Titus" [ref=e1591]
                - paragraph [ref=e1592]: Titus
                - generic [ref=e1593]:
                  - img [ref=e1594]
                  - generic [ref=e1596]: "6.9"
              - link "Planet of the Apes Planet of the Apes 6.9" [ref=e1597] [cursor=pointer]:
                - /url: /tv/19
                - img "Planet of the Apes" [ref=e1599]
                - paragraph [ref=e1600]: Planet of the Apes
                - generic [ref=e1601]:
                  - img [ref=e1602]
                  - generic [ref=e1604]: "6.9"
              - 'link "Mobile Suit Gundam: The Origin - Advent of the Red Comet Mobile Suit Gundam: The Origin - Advent of the Red Comet 8.4" [ref=e1605] [cursor=pointer]':
                - /url: /tv/88865
                - 'img "Mobile Suit Gundam: The Origin - Advent of the Red Comet" [ref=e1607]'
                - paragraph [ref=e1608]: "Mobile Suit Gundam: The Origin - Advent of the Red Comet"
                - generic [ref=e1609]:
                  - img [ref=e1610]
                  - generic [ref=e1612]: "8.4"
              - 'link "Heroes: Legend of Battle Disks Heroes: Legend of Battle Disks 5.8" [ref=e1613] [cursor=pointer]':
                - /url: /tv/88905
                - 'img "Heroes: Legend of Battle Disks" [ref=e1615]'
                - paragraph [ref=e1616]: "Heroes: Legend of Battle Disks"
                - generic [ref=e1617]:
                  - img [ref=e1618]
                  - generic [ref=e1620]: "5.8"
              - link "Rewriting Destiny Rewriting Destiny 2.0" [ref=e1621] [cursor=pointer]:
                - /url: /tv/202159
                - img "Rewriting Destiny" [ref=e1623]
                - paragraph [ref=e1624]: Rewriting Destiny
                - generic [ref=e1625]:
                  - img [ref=e1626]
                  - generic [ref=e1628]: "2.0"
              - link "Carnivàle Carnivàle 7.9" [ref=e1629] [cursor=pointer]:
                - /url: /tv/185
                - img "Carnivàle" [ref=e1631]
                - paragraph [ref=e1632]: Carnivàle
                - generic [ref=e1633]:
                  - img [ref=e1634]
                  - generic [ref=e1636]: "7.9"
              - link "Weeds Weeds 7.5" [ref=e1637] [cursor=pointer]:
                - /url: /tv/186
                - img "Weeds" [ref=e1639]
                - paragraph [ref=e1640]: Weeds
                - generic [ref=e1641]:
                  - img [ref=e1642]
                  - generic [ref=e1644]: "7.5"
              - link "Nine Perfect Strangers Nine Perfect Strangers 6.9" [ref=e1645] [cursor=pointer]:
                - /url: /tv/88989
                - img "Nine Perfect Strangers" [ref=e1647]
                - paragraph [ref=e1648]: Nine Perfect Strangers
                - generic [ref=e1649]:
                  - img [ref=e1650]
                  - generic [ref=e1652]: "6.9"
              - 'link "Crime Diaries: Night Out Crime Diaries: Night Out 7.3" [ref=e1653] [cursor=pointer]':
                - /url: /tv/89008
                - 'img "Crime Diaries: Night Out" [ref=e1655]'
                - paragraph [ref=e1656]: "Crime Diaries: Night Out"
                - generic [ref=e1657]:
                  - img [ref=e1658]
                  - generic [ref=e1660]: "7.3"
              - link "Tokyo 23-ku Onna Tokyo 23-ku Onna 5.0" [ref=e1661] [cursor=pointer]:
                - /url: /tv/89010
                - img "Tokyo 23-ku Onna" [ref=e1663]
                - paragraph [ref=e1664]: Tokyo 23-ku Onna
                - generic [ref=e1665]:
                  - img [ref=e1666]
                  - generic [ref=e1668]: "5.0"
```

# Test source

```ts
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
  172 |     await page.click('button:has-text("TV SHOWS")')
  173 |     await page.waitForTimeout(500)
  174 |     
  175 |     // Click first TV card
  176 |     await page.click('a[href^="/tv/"]')
  177 |     
  178 |     // Wait for TV detail API
  179 |     await waitForAPIResponse(page, '/api/discover/tv/')
  180 |     
  181 |     // Check TV detail elements
  182 |     await expect(page.locator('h1')).toBeVisible()
  183 |     await expect(page.locator('text=Overview')).toBeVisible()
> 184 |     await expect(page.locator('text=Episodes')).toBeVisible()
      |                                                 ^ Error: expect(locator).toBeVisible() failed
  185 |   })
  186 | 
  187 |   test('should show season selector', async ({ page }) => {
  188 |     await page.goto(`${BASE_URL}/tv/1399`) // Game of Thrones as example
  189 |     
  190 |     await waitForAPIResponse(page, '/api/discover/tv/')
  191 |     
  192 |     // Check season selector
  193 |     await expect(page.locator('select')).toBeVisible()
  194 |     await expect(page.locator('text=Show Info')).toBeVisible()
  195 |   })
  196 | })
  197 | 
  198 | test.describe('Media Manager - Search', () => {
  199 |   test('should search for movies', async ({ page }) => {
  200 |     await page.goto(`${BASE_URL}/search`)
  201 |     
  202 |     // Type search query
  203 |     await page.fill('input[type="text"]', 'Inception')
  204 |     
  205 |     // Submit search
  206 |     await page.keyboard.press('Enter')
  207 |     
  208 |     // Wait for search results
  209 |     await page.waitForTimeout(2000)
  210 |     
  211 |     // Check results loaded
  212 |     await expect(page.locator('text=Inception').first()).toBeVisible()
  213 |   })
  214 | })
  215 | 
  216 | test.describe('Media Manager - Responsive', () => {
  217 |   test('should be responsive on mobile', async ({ page }) => {
  218 |     await page.setViewportSize({ width: 375, height: 667 })
  219 |     await page.goto(`${BASE_URL}/discover`)
  220 |     
  221 |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  222 |     
  223 |     // Check that content is visible on mobile
  224 |     await expect(page.locator('h1:has-text("Discover")')).toBeVisible()
  225 |     
  226 |     // Check horizontal scrolling on movie rows
  227 |     const movieRow = page.locator('.overflow-x-auto').first()
  228 |     await expect(movieRow).toBeVisible()
  229 |   })
  230 | 
  231 |   test('should be responsive on tablet', async ({ page }) => {
  232 |     await page.setViewportSize({ width: 768, height: 1024 })
  233 |     await page.goto(`${BASE_URL}/discover`)
  234 |     
  235 |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  236 |     
  237 |     await expect(page.locator('h1:has-text("Discover")')).toBeVisible()
  238 |   })
  239 | })
  240 | 
  241 | test.describe('Media Manager - Accessibility', () => {
  242 |   test('should have proper heading structure', async ({ page }) => {
  243 |     await page.goto(`${BASE_URL}/discover`)
  244 |     
  245 |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  246 |     
  247 |     // Check for h1
  248 |     const h1 = await page.locator('h1').count()
  249 |     expect(h1).toBeGreaterThan(0)
  250 |     
  251 |     // Check for h2 sections
  252 |     const h2 = await page.locator('h2').count()
  253 |     expect(h2).toBeGreaterThan(0)
  254 |   })
  255 | 
  256 |   test('should have clickable navigation links', async ({ page }) => {
  257 |     await page.goto(BASE_URL)
  258 |     
  259 |     // Check all nav links are clickable
  260 |     const links = await page.locator('nav a').all()
  261 |     expect(links.length).toBeGreaterThan(0)
  262 |     
  263 |     for (const link of links) {
  264 |       await expect(link).toBeVisible()
  265 |     }
  266 |   })
  267 | })
  268 | 
```