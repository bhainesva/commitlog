import React, { useState, useEffect } from "react";
import PackagePicker from "./PackagePicker";
import ReactDiffViewer from 'react-diff-viewer';
import { DndProvider } from 'react-dnd'
import { HTML5Backend } from 'react-dnd-html5-backend'
import { DraggableList } from "./DraggableList"
import './scss/LandingPage.scss';
import './scss/Error.scss';
import * as R from 'ramda'

declare var Prism: any;

interface FileInfo {
  [key: string]: string
}

function ErrorToast(err: string) {
  if (!err) return null;

  return (
    <div className="Error">
      {err}
    </div>
  )
}

export default function App() {
  const [error, setError] = useState('');
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
      setPackages(data)
    })
  }, [])

  const fetchPackages = async () => {
    return fetch('http://localhost:3000/listPackages')
    .then(r => r.json())
  }

  const showErrorToast = (err: string) => {
    setError(err)
    setTimeout(() => setError(''), 3500)
  }


  const fetchFiles = async (pkg: string, testNames: string[], sortType: string) => {
    return fetch('http://localhost:3000/listFiles', {
      method: 'POST',
      body: JSON.stringify({
        pkg,
        tests: testNames,
        sort: sortType,
      })
    })
      .then(r => r.text())
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
    if (!R.contains(pkg, packages)) {
      showErrorToast(`Can't find package "${pkg}" please choose from the autocomplete!`)
    } else {
      const testNames = await fetchTestNames(pkg);
      setActivePkg(pkg)
      setTests(testNames);
      setFiles([])
    }
  }

  async function checkJobStatus(id: string) {
    const data = await fetch(`http://localhost:3000/job/${id}`)
      .then(r => r.json())

    if (data.Complete) {
      console.log("Job Complete! Updating files and tests")
      setTests(data.Results.tests);
      setActiveTest(0);
      setFiles(data.Results.files);
    } else {
      if (data.Error) {
        console.error("Job failed!: ", data.Error)
      } else {
        console.log("Current Status: ", data.Details)
        setTimeout(() => checkJobStatus(id), 300)
      }
    }
  }

  async function handleGenerateLogs(sortType: string) {
    fetchFiles(activePkg, tests, sortType).then(data => {
      checkJobStatus(data)
    })
  }

  // Landing Page
  if (tests.length === 0 && files.length === 0) {
    return (
      <div className="LandingPage">
        {ErrorToast(error)}
        {packages.length === 0 ?
          <div>Loading...</div> : <PackagePicker packages={packages} onSubmit={handleSubmit} />
        }
      </div>
    )
  }

  // Pkg selected, tests not ordered
  if (files.length === 0) {
    return (
      <div className="TestOrdering">
        <PackagePicker packages={packages} onSubmit={handleSubmit} />
        {error}
        <div>
        Choose an automatic test ordering
        <button onClick={() => handleGenerateLogs("raw")}>Generate with tests sorted by raw lines covered</button>
        <button onClick={() => handleGenerateLogs("net")}>Generate with tests sorted by net lines covered</button>
        <button onClick={() => handleGenerateLogs("importance")}>Generate with tests sorted by 'importance' heuristic</button>
        </div>

        Or manually order your tests
        <button onClick={() => handleGenerateLogs("")}>Generate with this order</button>
        <DndProvider backend={HTML5Backend}>
          <DraggableList setItems={setTests} items={tests} />
        </DndProvider>
      </div>
    )
  }

  return (
    <div>
      <PackagePicker packages={packages} onSubmit={handleSubmit} />
      {error}
      <br /><br />

      <div className="Page">
        <div className="Tests">
          <div>
            Tests
            ---------------
          </div>
          {tests.map((t, i) => <button key={i} name={t} className={i === activeTest ? 'is-active' : ''} onClick={() => setActiveTest(i)}>{t}</button>)}
          <button className={tests.length == activeTest ? 'is-active' : ''} onClick={() => setActiveTest(tests.length)}>Final</button>
        </div>
        <div className="Files">
          {filesView()}
        </div>
      </div>
    </div>
  )
}
