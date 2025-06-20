import { useEffect, useState } from "react";
import Setup from "./Setup";
import TeXBookLogoSM from "../components/LogoSM";
import { API_URL } from "../config";
import { Document, Page, pdfjs } from "react-pdf";
import Signal from "../components/Signal";

pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`;

type FileNode = {
  isDir: boolean
  name: string
  hash: string
  children: FileNode[]
}

type DataSource = {
  name: string;
  metadata: any;
};

export default function Home() {
  const [data, setData] = useState<DataSource[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedSource, setSelectedSource] = useState<string | null>(null);
  const [files, setFiles] = useState<FileNode | null>(null);
  const [filesLoading, setFilesLoading] = useState(false);
  const [filesError, setFilesError] = useState<string | null>(null);
  const [ping, setPing] = useState<number | null>(null) 

  useEffect(() => {
    async function fetchData() {
      try {
        const res = await fetch(`${API_URL}/api/v1/datasources/list`);
        if (!res.ok) throw new Error(`Error ${res.status}: ${res.statusText}`);
        const json = await res.json();
        setData(json);

        if (json.length > 0) {
          setSelectedSource(json[0].name || json[0].id || "");
        }
      } catch (err: any) {
        setError(err.message || "Unknown error");
      } finally {
        setLoading(false);
      }
    }

    fetchData();
  }, []);

  useEffect(() => {
    if (!selectedSource) {
      setFiles(null);
      return;
    }

    setFilesLoading(true);
    setFilesError(null);

    async function fetchFiles() {
      try {
        const res = await fetch(
          `${API_URL}/api/v1/datasources/${encodeURIComponent(selectedSource!)}/fs/list`
        );
        if (!res.ok) throw new Error(`Error ${res.status}: ${res.statusText}`);
        const json = await res.json();
        setFiles(json || []);
      } catch (err: any) {
        setFilesError(err.message || "Unknown error fetching files");
      } finally {
        setFilesLoading(false);
      }
    }

    fetchFiles();
  }, [selectedSource]);

  useEffect(() => {
    async function fetchPing() {
        const res = await fetch(
          `${API_URL}/api/v1/ping`
        );
        const json = await res.json();
        setPing(json.latency || 0)

    }

    fetchPing()
  }, [])

  if (loading) return <p>Loading data sources...</p>;
  if (error) return <p className="text-red-600">Error: {error}</p>;

  if (data && data.length === 0) {
    return <Setup />;
  }

  return (
    <div className="min-h-screen bg-white flex flex-col">
      <header className="w-full flex items-center justify-center relative px-4 py-4 border-b border-gray-200">
        <div className="absolute left-4">
          {ping != null && (
            <Signal delay={ping} />
          )}
          
        </div>
        <TeXBookLogoSM />
        <div className="absolute right-4 flex items-center gap-2">
          <label
            htmlFor="data-source-select"
            className="text-sm font-medium text-gray-700"
          >
            Source:
          </label>
          <select
            id="data-source-select"
            value={selectedSource ?? ""}
            onChange={(e) => setSelectedSource(e.target.value)}
            className="border border-gray-300 rounded-md p-1"
          >
            {data?.map((source) => (
              <option key={source.name} value={source.name}>
                {source.name}
              </option>
            ))}
          </select>
        </div>
      </header>

      <main className="flex-grow p-8 bg-gray-100 min-h-screen">
        {filesLoading && <p className="text-lg text-gray-700">Loading files...</p>}
        {filesError && <p className="text-red-600">Error: {filesError}</p>}

        {files && files.isDir && (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-6">
            {files.children
              ?.filter((child) => !child.isDir)
              .map((file) => {
                const encodedPath = encodeURIComponent("cache/" + file.hash + ".pdf");
                const fileUrl = `${API_URL}/api/v1/datasources/${encodeURIComponent(
                  selectedSource!
                )}/fs/file?path=${encodedPath}`;

                return (
                  <a
                    key={file.name}
                    href={"/viewer/" + selectedSource + "?path=" + "cache/" + file.hash+".pdf"}
                    className="group transition-transform duration-200 hover:shadow-xl rounded-2xl border border-gray-300 shadow-sm bg-white overflow-hidden"
                  >
                    <div className="aspect-[8.5/11] w-full bg-gray-50 flex items-center justify-center overflow-hidden">
                      <Document
                        file={fileUrl}
                        loading={
                          <div className="text-gray-400 text-sm">Loading preview…</div>
                        }
                        onLoadError={(err) => console.error("PDF Load Error:", err)}
                      >
                        <Page
                          pageNumber={1}
                          width={250}
                          scale={2}
                          renderTextLayer={false}
                          renderAnnotationLayer={false}
                          renderMode="canvas"
                        />
                      </Document>
                    </div>

                    <div className="p-3 border-t border-gray-200">
                      <p className="text-sm text-center text-gray-700 truncate">{file.name}</p>
                    </div>
                  </a>
                );
              })}
          </div>
        )}
      </main>



    </div>
  );
}