import React, { useState } from "react";

interface Props {
  onSubmit: (pkg: string) => void
}

export default function PackagePicker(props: Props) {
  const [value, setValue] = useState("");

  return (
    <div>
      <div>Pick a package</div> 
      <form onSubmit={(e) => {
          e.preventDefault();
          props.onSubmit(value)}
        }>
        <input type="text" value={value} onChange={e => setValue(e.target.value)} />
        <button>Go!</button>
      </form>
    </div>
  )
}