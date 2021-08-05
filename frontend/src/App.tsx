import React, { useState, useEffect } from "react";

import ReactDiffViewer from 'react-diff-viewer';
import { DndProvider } from 'react-dnd'
import { HTML5Backend } from 'react-dnd-html5-backend'
import { DraggableList } from "./DraggableList"
import './scss/LandingPage.scss';
import './scss/Toast.scss';
import Toast, {ToastProps} from './Toast';
import LandingPage from './LandingPage';
import Page from './Page';
import * as R from 'ramda'
import { FileMap, FetchFilesRequest_SortType } from "./gen/api";

declare var Prism: any;

interface FileInfo {
  [key: string]: string
}

export default function App() {
  const [toast, setToast] = useState<ToastProps>();
  const [loadingMessage, setLoadingMessage] = useState('Loading package list...');
  const [tests, setTests] = useState([]);
  const [packages, setPackages] = useState([]);
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [collapseStatus, setCollapseStatus] = useState<{[key: string]: boolean}>({});
  const [activeTest, setActiveTest] = useState(-1);
  const [activePkg, setActivePkg] = useState("");

  const fetchTestNames = async (pkg: string) => {
    return fetch('http://localhost:3000/listTests?pkg=' + pkg)
      .then(r => r.json())
  }

  useEffect(() => {
    fetchPackages().then((data) => {
      setPackages(data);
      setLoadingMessage('');
    })
  }, [])

  const fetchPackages = async () => {
    return fetch('http://localhost:3000/listPackages')
    .then(r => r.json())
  }

  const showErrorToast = (msg: string) => {
    setToast({
      msg: msg,
      err: true,
    });
    setTimeout(() => setToast(null), 3500)
  }

  const showSuccessToast = (msg: string) => {
    setToast({
      msg: msg,
      err: false,
    })
    setTimeout(() => setToast(null), 3500)
  }


  const fetchFiles = async (pkg: string, testNames: string[], sortType: FetchFilesRequest_SortType) => {
    return fetch('http://localhost:3000/job', {
      method: 'POST',
      body: JSON.stringify({
        pkg,
        tests: testNames,
        sort: sortType,
      })
    })
      .then(r => r.json())
  }

  const toggleCollapse = (file: string) => {
    setCollapseStatus(R.assoc(file, !collapseStatus[file], collapseStatus))
  }

  const highlightSyntax = (str: string) => (
    <pre
      style={{ display: 'inline' }}
      dangerouslySetInnerHTML={{
        __html: Prism.highlight(str || "", Prism.languages.clike, 'javascript'),
      }}
    />
  );

  const filesView = () => {
    const previousContents = activeTest === 0 ? {} : files[activeTest - 1];
    const contents = files[activeTest];

    const out = [];
    for (const [name, content] of Object.entries(contents)) {
      const previousContent = previousContents[name] || '';
      out.push(
        <div key={name} className="File">
          <button onClick={() => toggleCollapse(name)} className="File-name">{name}</button>
          {collapseStatus[name] || 
          <ReactDiffViewer
            oldValue={atob(previousContent)}
            newValue={atob(content)}
            splitView={false}
            leftTitle={name}
            styles={{titleBlock: {display: "none"}}}
            showDiffOnly={Object.keys(contents).length > 1 ? true : false}
            renderContent={highlightSyntax}
          />}
        </div>
      )
    }

    return out
  }


  async function handleSubmit(pkg: string) {
    if (!R.contains(pkg, packages) && !pkg.startsWith("/")) {
      showErrorToast(`Can't find package "${pkg}" please choose from the autocomplete, or provide an absolute path`)
    } else {
      setLoadingMessage("Fetching tests...")
      const testNames = await fetchTestNames(pkg);
      if (!testNames) {
        showErrorToast("That package has no tests!");
        setLoadingMessage("")
        return;
      }
      setLoadingMessage("")
      setActivePkg(pkg)
      setTests(testNames);
      setFiles([])
    }
  }

  function checkoutFiles() {
    const data = fetch(`http://localhost:3000/checkout`, {
      method: 'POST',
      body: JSON.stringify({
        files: files[activeTest],
      })
    })
    .then(r => console.log(r.status))
  }

  async function checkJobStatus(id: string) {
    const data = await fetch(`http://localhost:3000/job/${id}`)
      .then(r => r.json())

    if (data.complete) {
      showSuccessToast("Processing Finished!")
      setTests(data.results.tests);
      setActiveTest(0);
      setLoadingMessage('')
      setFiles(data.results.files.map((x: FileMap) => x.files));
    } else {
      if (data.Error) {
        showErrorToast("Job failed!: " + data.error)
        console.error("Job failed!: ", data.error)
      } else {
        setLoadingMessage(data.details + '...')
        setTimeout(() => checkJobStatus(id), 300)
      }
    }
  }

  async function handleGenerateLogs(sortType: FetchFilesRequest_SortType) {
    fetchFiles(activePkg, tests, sortType).then(data => {
      setLoadingMessage('Analyzing package...')
      checkJobStatus(data.id)
    })
  }

  const pageContent = () => {
    if (loadingMessage) {
      return (
        <div className="LandingPage">
          <div>
            {loadingMessage}
          </div>
        </div>
      )
    }

    if (tests.length === 0 && files.length === 0) {
      return (
        <LandingPage packages={packages} onSubmit={handleSubmit}/>
      )
    }

    if (files.length === 0) {
      return (
        <div className="TestOrdering">
          <div className="TestOrdering-auto">
            <h2>Choose an automatic test ordering</h2>
            <div>
              <button onClick={() => handleGenerateLogs(FetchFilesRequest_SortType.RAW)}>Generate with tests sorted by raw lines covered</button>
              <button onClick={() => handleGenerateLogs(FetchFilesRequest_SortType.IMPORTANCE)}>Generate with tests sorted by net lines covered</button>
              <button onClick={() => handleGenerateLogs(FetchFilesRequest_SortType.NET)}>Generate with tests sorted by 'importance' heuristic*</button>
            </div>

            <div>*The heuristic works by computing a score for each line of the code, based on the number of tests that cover it. Tests are then ordered by the average score of the lines they cover</div>
          </div>

          <div className="TestOrdering-manual">
            <h2>Or manually order your tests (click and drag to reorder)</h2>
            <button onClick={() => handleGenerateLogs(FetchFilesRequest_SortType.HARDCODED)}>Generate with this order</button>
            <DndProvider backend={HTML5Backend}>
              <DraggableList setItems={setTests} items={tests} />
            </DndProvider>
          </div>
        </div>
      )
    }

    return (
      <>
        <div className="TestBrowser">
          <div className="TestBrowser-tests">
            {tests.map((t, i) => <button key={i} name={t} className={i === activeTest ? 'is-active' : ''} onClick={() => setActiveTest(i)}>{t}</button>)}
            <button className={tests.length == activeTest ? 'is-active' : ''} onClick={() => setActiveTest(tests.length)}>Final</button>

            <button onClick={checkoutFiles}>Checkout Files</button>
          </div>
          <div className="TestBrowser-files">
            {filesView()}
          </div>
        </div>
      </>
    )
  }

  return (
    <Page hidePicker={!activePkg} packages={packages} activePackage={activePkg} onSubmit={handleSubmit}>
      {Toast(toast)}
      {pageContent()}
    </Page>
  )
}
