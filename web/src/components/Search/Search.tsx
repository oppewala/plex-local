import React, {ChangeEventHandler, useMemo, useState} from "react";
import {Search as SearchApi, Download as DownloadApi, DownloadPersist as DownloadPersistApi} from '@services/Api/Api.service';
import {DownloadPersistResponse, DownloadResponse, SearchResponse} from "@services/Api/types";
import {throttle} from "lodash";

export const Search = () => {
    const [query, setQuery] = useState('');
    const [results, setResults] = useState<SearchResponse[]>([]);

    const setOrderedResults = (search: string, results: SearchResponse[]) => {
        search = search.toLowerCase()

        // TODO: Do sorting better in API
        const ordered = results
            .filter((r) => {
                if (r.LowercaseTitle.startsWith(search)) {
                    return true
                }
                if (r.LowercaseTitle.includes(search)) {
                    return true
                }
                return r.Similarity > 0.4;
            })
            .sort(((a, b) => {
                if (a.LowercaseTitle.startsWith(search) && b.LowercaseTitle.startsWith(search)) {
                    return b.Similarity - a.Similarity;
                }

                if (a.LowercaseTitle.startsWith(search)) {
                    return -1;
                }

                if (b.LowercaseTitle.startsWith(search)) {
                    return 1;
                }

                return b.Similarity - a.Similarity;
            }))

        setResults(ordered);
    }

    const search = useMemo(() => throttle(SearchApi, 400), []);

    const changeHandler: ChangeEventHandler<HTMLInputElement> = async (e) => {
        const q = e.target.value;
        setQuery(q)
        search(q)!
            .then(r => setOrderedResults(q, r))
    }

    return (
        <form autoComplete='false' className='w-full relative'>
            <input type='text' id='search' name='search' onChange={changeHandler} value={query} placeholder='Search...'
                   autoComplete='off'
                   className='text-2xl w-full px-6 py-4 bg-white rounded-xl shadow-md space-x-4 focus:outline-none focus:ring focus:border-blue-100'/>
            <ResetButton display={query !== ''} resetFunction={() => {
                setQuery('');
                setResults([])
            }}/>
            <SearchResults results={results}/>
        </form>)
}

type ResetFunction = () => void;
const ResetButton: React.FC<{ display: boolean, resetFunction: ResetFunction }> = ({display, resetFunction}) => {
    if (!display) return null;

    return (
        <span onClick={resetFunction}
              className='absolute inset-y-0 right-0 pr-6 pt-5 flex cursor-pointer text-md h-16'>
                🗙
        </span>
    )
}

const SearchResult: React.FC<{ result: SearchResponse }> = ({result}) => {
    const download = (key: string) => {
        DownloadApi(key)
            .then((r: DownloadResponse) => console.log('download', r))
    }
    const downloadPersist = (key: string) => {
        DownloadPersistApi(key)
            .then((r: DownloadPersistResponse) => console.log('download', r))
    }
    const typeIcon = (type: string) => {
        if (type === 'movie') {
            return '🎬'
        }

        if (type === 'show') {
            return '📺'
        }

        return type;
    }
    let tooltips: {
        download: {
            [key: string]: string
        },
        downloadPersist: {
            [key: string]: string
        }
    }
    tooltips = {
        download: {
            show: 'Download all episodes currently available on plex',
            movie: 'Download Movie'
        },
        downloadPersist: {
            show: 'Download future episodes as they release'
        }
    }

    return (<li className='my-3 px-6 py-3 bg-blue-100 rounded-xl shadow-md space-x-4 flex flex-row'>
        <div className='flex-1'>{typeIcon(result.Type)} {result.Title}</div>
        <div className='flex flex-row'>
            <span onClick={() => download(result.Key)} className='tt tt-top flex-none cursor-pointer' data-text={tooltips.download[result.Type]}>🔽</span>
            {result.Type === 'show' ?
                <span onClick={() => downloadPersist(result.Key)} className='tt tt-top flex-none cursor-pointer' data-text={tooltips.downloadPersist[result.Type]}>⏬</span>
                : null}
        </div>
    </li>)
}

const SearchResults: React.FC<{ results: Array<SearchResponse> }> = ({results}) => {
    return (<div>
        <ul>
            {results.map(r => <SearchResult key={r.Key} result={r}/>)}
        </ul>
    </div>)
}

export default Search;