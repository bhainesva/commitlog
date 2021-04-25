import React, { useState } from "react";
import PackagePicker from "./PackagePicker";
import ReactDiffViewer from 'react-diff-viewer';

declare var Prism: any;

interface FileInfo {
  [key: string]: string
}

export default function App() {
  const [tests, setTests] = useState([]);
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [activeTest, setActiveTest] = useState(-1);

  const fetchTestNames = async (pkg: string) => {
    return fetch('http://localhost:3000/listTests?pkg=' + pkg)
      .then(r => r.json())
  }

  const fetchFiles = async (pkg: string, testNames: string[]) => {
    return fetch('http://localhost:3000/listFiles', {
      method: 'POST',
      body: JSON.stringify({
        pkg,
        tests: testNames,
      })
    })
      .then(r => r.json())
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
    console.log("Active: ", activeTest);
    const previousContents = activeTest === 0 ? {} : files[activeTest - 1];
    const contents = files[activeTest];
    console.log("contenst", contents);

    const out = [];
    for (const [name, content] of Object.entries(contents)) {
      const previousContent = previousContents[name] || '';
      console.log("prev content: ", previousContent);
      console.log("new content", content);
      out.push(
        <div className="File">
          <ReactDiffViewer
            oldValue={atob(previousContent)}
            newValue={atob(content)}
            splitView={false}
            leftTitle={name}
            showDiffOnly={Object.keys(contents).length > 1 ? true : false}
            renderContent={highlightSyntax}
          />
        </div>
      )
    }

    return out
  }

  async function handleSubmit(pkg: string) {
    const testNames = await fetchTestNames(pkg);
    fetchFiles(pkg, testNames)
      .then((data) => {
        setFiles(data)
        setTests(testNames)
      })
  }

  return (
    <div>
      <PackagePicker onSubmit={handleSubmit} />
      <br /><br />
      <div className="Page">
        <div className="Tests">
          <div>
            Tests
            ---------------
          </div>
          {tests.map((t, i) => <button className={i == activeTest ? 'is-active' : ''} onClick={() => setActiveTest(i)} key={i}>{t}</button>)}
          <button className={tests.length == activeTest ? 'is-active' : ''}  onClick={() => setActiveTest(tests.length)}>Final</button>
        </div>
        <div className="Files">
          {filesView()}
        </div>
      </div>
    </div>
  )
}