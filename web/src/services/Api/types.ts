export type SearchResponse = {
    Type: string;
    Key: string;
    Title: string;
    LowercaseTitle: string;
    ParentTitle: string;
    GrandparentTitle: string;
    Similarity: number;
}

export type DownloadResponse = {
    Message: string;
}

export type DownloadPersistResponse = {
    Message: string
}