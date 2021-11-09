import React, {useContext, useMemo, useState} from "react";
import SocketContext, {Message} from "@components/SocketContext/SocketContext";

interface DownloadUpdateMessage extends Message {
    Title: string
    BytesDownloaded: number
    TotalBytes: number
}

type Download = {
    Title: string
    BytesDownloaded: number
    TotalBytes: number
    Complete: boolean
}

export const Downloads = () => {
    const [downloads, setDownloads] = useState<{ [title: string]: Download }>({})

    const socketContext = useContext(SocketContext);

    const handleDownloadMessage = (message: Message) => {
        console.log(message)

        const msg = message as DownloadUpdateMessage;
        setDownloads((prevState: { [title: string]: Download }) => {
            let newState = {
                ...prevState
            };
            newState[msg.Title] = {
                Title: msg.Title,
                BytesDownloaded: msg.BytesDownloaded,
                TotalBytes: msg.TotalBytes,
                Complete: false
            };
            console.log('setting state', msg, msg.Title, prevState, newState)
            return newState;
        });
    }

    const handleCompleteMessage = (message: Message) => {
        const msg = message as DownloadUpdateMessage;
        setDownloads((prevState: { [title: string]: Download }) => {
            let newState = {
                ...prevState
            };
            newState[msg.Title] = {
                Title: msg.Title,
                BytesDownloaded: 100,
                TotalBytes: 100,
                Complete: true
            };
            return newState;
        });
    }

    const handleStartMessage = (message: Message) => {
        const msg = message as DownloadUpdateMessage;
        setDownloads((prevState: { [title: string]: Download }) => {
            let newState = {
                ...prevState
            };
            newState[msg.Title] = {
                Title: msg.Title,
                BytesDownloaded: msg.BytesDownloaded,
                TotalBytes: msg.TotalBytes,
                Complete: false
            };
            return newState;
        });
    }

    useMemo(() => {
        socketContext.AddListener({
            Type: 'open',
            Emit: (data : Event) => {
                console.log('connection opened', data)
            }
        })
        socketContext.AddListener({
            Type: 'close',
            Emit: (data : Event) => {
                // TODO: Notify and/or reconnect
                console.log('connection closed', data)
            }
        })
        socketContext.AddListener({
            MessageType: 'download-update',
            Emit: handleDownloadMessage
        })
        socketContext.AddListener({
            MessageType: 'download-start',
            Emit: handleStartMessage
        })
        socketContext.AddListener({
            MessageType: 'download-complete',
            Emit: handleCompleteMessage
        })
    }, [socketContext])

    return (
        <div className='w-full'>
            <ul className='w-full'>{Object.getOwnPropertyNames(downloads).map((t) => <DownloadBar key={t} title={t}
                                                                                                  downloads={downloads}/>)}</ul>
        </div>
    )
}

const DownloadBar: React.FC<{ title: string, downloads: { [title: string]: Download } }> = ({title, downloads}) => {
    const download = downloads[title];

    const percentNumber = (dl: Download): number => {
        if (dl.Complete) return 100;
        if (dl.BytesDownloaded === 0) return 0;

        const p = dl.BytesDownloaded / dl.TotalBytes * 100;
        return Math.round((p + Number.EPSILON) * 10) / 10
    }

    const perc = percentNumber(download)
    const progressStyle = {
        width: perc + '%'
    }
    const animateStyle = !download.Complete ? 'animate-pulse' : '';

    return (
        <li className='relative grid-cols-1 text-md my-6 overflow-hidden rounded-xl shadow-md px-6 py-3'>
            <div>
                <div className='-z-10 absolute inset-0 w-full h-full bg-green-100' />
                <div
                    className={`-z-10 absolute inset-0 h-full bg-gradient-to-r from-green-400 to-green-500 ${animateStyle}`}
                    style={progressStyle} />
                {(!download.Complete && download.BytesDownloaded !== 0) && <div
                  className={`-z-10 absolute inset-0 h-full bg-none border-solid border-r border-green-600 ${animateStyle}`}
                  style={progressStyle} />}
            </div>
            <div className=''>{title} {perc === 0 ? '(queued)' : `(${perc}%)`}</div>
        </li>)
}

export default Downloads;