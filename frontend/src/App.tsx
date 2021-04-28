import React, { useState, useEffect } from "react";
import PackagePicker from "./PackagePicker";
import ReactDiffViewer from 'react-diff-viewer';
import { DndProvider } from 'react-dnd'
import { HTML5Backend } from 'react-dnd-html5-backend'
import { DraggableList } from "./DraggableList"
import * as R from 'ramda'

declare var Prism: any;

interface FileInfo {
  [key: string]: string
}

export const ItemTypes = {
  TEST: 'test'
}

export default function App() {
  const [errorState, setErrorState] = useState('');
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


  const fetchFiles = async (pkg: string, testNames: string[], sortType: string) => {
    return fetch('http://localhost:3000/listFiles', {
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
    if (activeTest == -1) return 'Select a test to begin';
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
      setErrorState('We don\'t recognize that package!');
      setTimeout(() => setErrorState(''), 3000);
    } else {
      const testNames = await fetchTestNames(pkg);
      setActivePkg(pkg)
      setTests(testNames);
      setFiles([])
    }
  }

  async function handleGenerateLogs(sortType: string) {
    fetchFiles(activePkg, tests, sortType).then(data => {
      console.log("got: ", data);
      setTests(data.tests);
      setFiles(data.files);
    })
  }

  // Landing Page
  if (tests.length === 0 && files.length === 0) {
    return (
      <div className="LandingPage">
        {packages.length === 0 ?
          <div>Loading...</div> : <PackagePicker packages={packages} onSubmit={handleSubmit} />
        }
        {errorState}
      </div>
    )
  }

  // Pkg selected, tests not ordered
  if (files.length === 0) {
    return (
      <div className="TestOrdering">
        <PackagePicker packages={packages} onSubmit={handleSubmit} />
        {errorState}
        <div>
          Order your tests man
        </div>
        <DndProvider backend={HTML5Backend}>
          <DraggableList setItems={setTests} items={tests} />
        </DndProvider>
        <button onClick={() => handleGenerateLogs("")}>Generate with this order</button>
        <button onClick={() => handleGenerateLogs("raw")}>Generate with tests sorted by raw lines covered</button>
        <button onClick={() => handleGenerateLogs("net")}>Generate with tests sorted by net lines covered</button>
        <button onClick={() => handleGenerateLogs("importance")}>Generate with tests sorted by 'importance' heuristic</button>
      </div>
    )
  }

  return (
    <div>
      <PackagePicker packages={packages} onSubmit={handleSubmit} />
      {errorState}
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
