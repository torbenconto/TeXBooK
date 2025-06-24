import { useEffect, useState } from "react";
import Setup from "./Setup";
import TeXBookLogoSM from "../components/LogoSM";
import { API_URL } from "../config";
import { Document, Page, pdfjs } from "react-pdf";
import Signal from "../components/Signal";
import { useSearchParams } from "react-router-dom";

pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`;

type FileNode = {
  isDir: boolean;
  name: string;
  hash: string;
  children: FileNode[];
};

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
  const [ping, setPing] = useState<number | null>(null);
  const [currentPath, setCurrentPath] = useState<string[]>([]);
  const [searchParams, setSearchParams] = useSearchParams();
  const [showAddSource, setShowAddSource] = useState(false);
  const [newSourceName, setNewSourceName] = useState("");
  const [addError, setAddError] = useState<string | null>(null);
  const [newSourcePath, setNewSourcePath] = useState("");
  const [searchBarContent, setSearchBarContent] = useState("");

  useEffect(() => {
    const pathParam = searchParams.get("path");
    if (pathParam) {
      const segments = pathParam.split("/").filter(Boolean);
      setCurrentPath(segments);
    }
  }, []);

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
      const MIN_LOAD_DURATION = 300;
      const start = performance.now();

      try {
        const joinedPath = currentPath.join("/");
        searchParams.set("path", joinedPath);
        setSearchParams(searchParams);

        const res = await fetch(
          `${API_URL}/api/v1/datasources/${encodeURIComponent(selectedSource!)}/fs/list?path=${encodeURIComponent(joinedPath)}`
        );
        if (!res.ok) throw new Error(`Error ${res.status}: ${res.statusText}`);
        const json = await res.json();
        setFiles(json || []);

        const elapsed = performance.now() - start;
        const delay = Math.max(0, MIN_LOAD_DURATION - elapsed);
        setTimeout(() => {
          setFilesLoading(false);
        }, delay);
      } catch (err: any) {
        setFilesError(err.message || "Unknown error fetching files");
        setFilesLoading(false);
      }
    }

    fetchFiles();
  }, [selectedSource, currentPath]);

  useEffect(() => {
    async function fetchPing() {
      const start = performance.now();
      try {
        const res = await fetch(`${API_URL}/api/v1/ping`);
        await res.json();
        const end = performance.now();
        setPing(end - start);
      } catch {
        setPing(null);
      }
    }

    const interval = setInterval(fetchPing, 5000);
    fetchPing();

    return () => clearInterval(interval);
  }, []);

  if (loading) return <p>Loading data sources...</p>;
  if (error) return <p className="text-red-600">Error: {error}</p>;
  if (data && data.length === 0) return <Setup />;

  return (
    <div className="min-h-screen bg-white flex flex-col">
      <header className="w-full flex items-center justify-center relative px-4 py-4 border-b border-gray-200">
        <div className="absolute left-4">
          {ping != null && <Signal delay={ping} />}
        </div>
        <TeXBookLogoSM />
        <div className="absolute right-4 flex items-center gap-2">
          <label htmlFor="data-source-select" className="text-sm font-medium text-gray-700">
            Source:
          </label>
          <select
            id="data-source-select"
            value={selectedSource ?? ""}
            onChange={(e) => {
              setCurrentPath([])
              searchParams.set("path", "")
              setSearchParams(searchParams)
              setSelectedSource(e.target.value)
            }}
            className="border border-gray-300 rounded-md p-1"
          >
            {data?.map((source) => (
              <option key={source.name} value={source.name}>
                {source.name}
              </option>
            ))}
          </select>
          <button
            onClick={() => {
              setNewSourceName("");
              setAddError(null);
              setShowAddSource(true);
            }}
            className="ml-2 px-2 py-1 bg-green-600 text-white text-sm rounded hover:bg-green-700 transition-colors"
          >
            + Add Source
          </button>
        </div>
      </header>

      <main className="flex-grow p-8 bg-gray-100 transition-all duration-300">
        <div className="flex w-full gap-2">
        {/* Breadcrumb path display */}
          <div className="flex items-center flex-wrap w-full bg-gray-100 p-2 border border-gray-300 rounded-lg mb-4 shadow-sm text-sm">
            <button
              onClick={() => {
                setCurrentPath([])
              }}
              className="text-[#098842] text-md hover:underline font-medium transition-colors"
            >
              root
            </button>
            <span className="mx-2 text-gray-500">/</span>
            {currentPath.map((elem: string, idx: number) => (
              <div key={idx} className="flex items-center">
                <button
                  onClick={() => {
                    setCurrentPath(currentPath.slice(0, idx + 1))
                  }}
                  className="text-[#098842] text-md hover:underline font-medium transition-colors"
                >
                  {elem}
                </button>
                {idx < currentPath.length - 1 && (
                  <span className="mx-2 text-gray-500">/</span>
                )}
              </div>
            ))}
          </div>        
          {/* Search bar inline with path view */}
          <div className="flex justify-end mb-4">
            <input
              type="text"
              placeholder="Search files..."
              value={searchBarContent}
              onChange={(e) => setSearchBarContent(e.target.value)}
              className="w-full sm:w-64 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none transition"
            />
          </div>
        </div>
        {filesError && <p className="text-red-600">Error: {filesError}</p>}

        {filesLoading && (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-6">
            {Array.from({ length: 8 }).map((_, idx) => (
              <div
                key={idx}
                className="animate-pulse rounded-2xl border border-gray-300 bg-white p-6 aspect-[8.5/11] flex flex-col items-center justify-center"
              >
                <div className="bg-gray-200 rounded-md w-12 h-12 mb-4" />
                <div className="h-4 bg-gray-200 rounded w-3/4" />
              </div>
            ))}
          </div>
        )}

        <div
          className={`grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-6 transition-opacity duration-300 ${
            filesLoading ? "opacity-0 pointer-events-none" : "opacity-100"
          }`}
        >
          {files &&
            !filesLoading &&
            files.isDir &&
            [...files.children]
              .sort((a, b) => {
                if (a.isDir && !b.isDir) return -1;
                if (!a.isDir && b.isDir) return 1;
                return a.name.toLowerCase().localeCompare(b.name.toLowerCase());
              })
              .map((file) => {
                const searchTerms = searchBarContent
                .trim()
                .toLowerCase()
                .split(/\s+/)
                .filter(term => term.length > 0);

              const matchesSearch =
                searchTerms.length === 0 ||
                searchTerms.every(term => file.name.toLowerCase().includes(term));
                const encodedPath = encodeURIComponent(
                  currentPath.join("/") + "/cache/" + file.hash + ".pdf"
                );
                const fileUrl = `${API_URL}/api/v1/datasources/${encodeURIComponent(
                  selectedSource!
                )}/fs/file?path=${encodedPath}`;

                
                if (file.isDir) {
                  return (
                    <button
                      key={file.name}
                      onClick={() => {
                        setFilesLoading(true);
                        setTimeout(() => setCurrentPath([...currentPath, file.name]), 50);
                      }}
                      className={`group transition-transform duration-200 hover:shadow-xl min-h-full min-w-full rounded-2xl border border-gray-300 shadow-sm bg-white overflow-hidden flex flex-col items-center justify-center p-6 cursor-pointer ${
                          matchesSearch ? "" : "hidden"
                        }`}                      
                      title={file.name}
                    >
                      <div className="text-gray-600 mb-3">
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          fill="#098842"
                          className="w-12 h-12"
                          viewBox="0 0 24 24"
                        >
                          <path d="M19.5 21a3 3 0 0 0 3-3v-4.5a3 3 0 0 0-3-3h-15a3 3 0 0 0-3 3V18a3 3 0 0 0 3 3h15ZM1.5 10.146V6a3 3 0 0 1 3-3h5.379a2.25 2.25 0 0 1 1.59.659l2.122 2.121c.14.141.331.22.53.22H19.5a3 3 0 0 1 3 3v1.146A4.483 4.483 0 0 0 19.5 9h-15a4.483 4.483 0 0 0-3 1.146Z" />
                        </svg>
                      </div>
                      <p className="text-sm text-gray-700 text-center truncate">{file.name}</p>
                    </button>
                  );
                } else {
                  if (!file.name || !file.hash) return null;

                  return (
                    <a
                      key={file.name}
                      href={`/viewer/${selectedSource}?path=${encodedPath}`}
                      className={`group transition-transform duration-200 hover:shadow-xl rounded-2xl border border-gray-300 shadow-sm bg-white overflow-hidden ${
                        matchesSearch ? "" : "hidden"
                      }`}
                    >
                      <div className="aspect-[8.5/11] w-full bg-gray-50 flex items-center justify-center overflow-hidden">
                      {fileUrl && (
                        <Document
                          file={fileUrl}
                          key={fileUrl}
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
                            className="transition-opacity duration-300"
                          />
                        </Document>
                      )}
                      </div>
                      <div className="p-3 border-t border-gray-200">
                        <p className="text-sm text-center text-gray-700 truncate">{file.name}</p>
                      </div>
                    </a>
                  );
                }
              })}
        </div>
      </main>
      
      {showAddSource && (
        <div className="fixed inset-0 bg-white/30 backdrop-blur-md flex justify-center items-center z-50">
          <div className="bg-white p-6 rounded-sm shadow-lg w-full max-w-md">
            <h2 className="text-lg font-semibold mb-4">Add New Source</h2>
            <input
              type="text"
              placeholder="Source name"
              value={newSourceName}
              onChange={(e) => setNewSourceName(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 mb-2"
            />
            <input
              type="text"
              placeholder="Path (e.g., /mnt/data/docs)"
              value={newSourcePath}
              onChange={(e) => setNewSourcePath(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 mb-4"
            />
            {addError && <p className="text-red-600 text-sm mb-2">{addError}</p>}
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setShowAddSource(false)}
                className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
              >
                Cancel
              </button>
              <button
                onClick={async () => {
                  setAddError(null);
                  try {
                    const res = await fetch(`${API_URL}/api/v1/datasources/add`, {
                      method: "POST",
                      headers: { "Content-Type": "application/json" },
                      body: JSON.stringify({ name: newSourceName, path: newSourcePath, type: "local" }),
                    });
                    if (!res.ok) {
                      const text = await res.text();
                      throw new Error(text || `Failed to add source`);
                    }

                    setShowAddSource(false);
                    setNewSourceName("");

                    const refetch = await fetch(`${API_URL}/api/v1/datasources/list`);
                    const json = await refetch.json();
                    setData(json);
                    setSelectedSource(json[json.length - 1]?.name || json[0]?.name || "");
                  } catch (err: any) {
                    setAddError(err.message || "Failed to add");
                  }
                }}
                className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
              >
                Add
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
