# package main

from dataclasses import dataclass
import asyncio, argparse, json, io, logging, os, os.path, string, time
import websockets  # pip3 install websockets

@dataclass
class Challenges:
	Msg: str
	# Values: List[Challenge]
	IsLabAssigned: bool

@dataclass
class Challenge:
	# ...
	IsUserCompleted: bool
	# TeamsCompleted
	IsChalDisabled: bool

# done chan interface{}
# interrupt chan os.Signal
haaukinsURL = ""
haaukinsSessionCookie = ""
filePath = ""

challenges: Challenges = []

cookieJar = "cookiejar.Jar"
haaukinsDialer: "websocket.Dialer"

async def init():
	global haaukinsURL, haaukinsDialer, filePath
	flag = argparse.ArgumentParser()
	flag.add_argument("--url", dest="haaukinsURL", default="http://localhost", help="Haaukins url fx: ddc.haaukins.com")
	flag.add_argument("--session", dest="haaukinsSessionCookie", default="", help="Haaukins session cookie")
	flag.add_argument("--file", dest="filePath", default="challenges", help="Path to save challenges, default is local subdirectory challenges")

	args = flag.parse_args()
	haaukinsURL = args.haaukinsURL
	filePath = args.filePath  # needed?
	# done = make(chan interface{})		  //Channel to indicate when the connection is closed
	# interrupt = make(chan os.Signal, 1)	// Channel to listen for interrupts
	# signal.Notify(interrupt, os.Interrupt) // Catch OS interrupts

	# cookieJar, err = cookiejar.New(None)
	# if err != None: {
	# 	exit(err)
	# }
	cookie = {
		"Name": "session",
		"Value": args.haaukinsSessionCookie
	}
	urlObj = "https://" + args.haaukinsURL
	# cookieJar.SetCookies(urlObj, []*http.Cookie{cookie})
	# haaukinsDialer = websocket.Dialer{
	# 	Jar: cookieJar,
	# }


async def main():
	global haaukinsURL
	socketURL = "wss://" + haaukinsURL + "/challengesFrontend"
	conn, resp, err = await websockets.connect(socketURL), "TODO", None
	# conn, resp, err = haaukinsDialer.Dial(socketURL, None)
	if err == websockets.ProtocolError: {
		logging.debug("Bad handshake: %d", resp.StatusCode)
	}
	if err != None: {
		exit(err)
	}
	# defer conn.Close()
	await recieveChallenges(conn)

	path = os.path.join(".", filePath)
	os.makedirs(path, exist_ok=True)  # os.ModePerm

	# var wg sync.WaitGroup
	# wg.Add(len(challenges.Values))
	logging.debug("Downloading files...")
	for chal in challenges.Values:
		handleChallenge(chal, path)
		downloadFileIfExists(chal, path)

	#wg.Wait()
	logging.debug("Done! Files saved to:", path)


def downloadFileIfExists(chal: Challenge, rootPath: str):
	fileUrl = "xurls.Strict().FindString(chal.Challenge.TeamDescription)"
	#client = &http.Client{}

	fsPath = os.path.join(rootPath, chal.Challenge.Tag)
	os.makedirs(fsPath, os.ModePerm)

	# Check for false positives
	if fileUrl != "" and not "hkn" in fileUrl:
		# req, err = http.NewRequest("GET", fileUrl, None)
		if err != None: {
			exit(err)
		}
		resp, err = client.Do(req)
		if err != None: {
			exit(err)
		}
		# defer resp.Body.Close()
		filePath = path.Base(req.URL.Path)
		if filePath == "download":
			filePath = chal.Challenge.Tag

		logging.debug("Downloading file:", filePath)
		file = os.Create(os.path.join(fsPath, filePath))
		if err != None: {
			exit(err)
		}
		# defer file.Close()
		_, err = io.Copy(file, resp.Body)
		if err != None: {
			exit(err)
		}


def handleChallenge(chal: Challenge, rootPath: str):
	path = os.path.join(rootPath, chal.Challenge.Tag)
	os.makedirs(path, exist_ok=True)  # os.ModePerm
	# 
	file, err = os.Create(os.path.join(path, "challenge.json"))
	if err != None: {
		exit(err)
	}
	# defer file.Close()
	with open(file, 'wt') as enc:
		json.dump(chal, file, indent=4)


async def recieveChallenges(conn: "websockets.legacy.client.Connect"):
	global challenges
	_, challengesJSON, err = "", await conn.recv(), None

	if err != None:
		print(err)
		return

	challenges = Challenges(**json.loads(challengesJSON))

asyncio.run(init())
asyncio.run(main())
