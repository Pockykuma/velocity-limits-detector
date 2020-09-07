import * as React from 'react'
import axios from 'axios'
import CSS from 'csstype'

const textareaStyles: CSS.Properties = {
  backgroundColor: 'lightgrey',
  boxShadow: '0 0 10px rgba(0, 0, 0, 0.3)',
  width: '600px',
  height: '500px',
}

export default class FileInput extends React.Component<{}, { output: string }> {
  constructor(props) {
    super(props)
    this.state = { output: '' }
    this.uploadFile = this.uploadFile.bind(this)
  }

  uploadFile = async (event: React.ChangeEvent<HTMLInputElement>) => {
    let file: File = event.target.files[0]

    if (file) {
      let data: FormData = new FormData()
      data.append('file', file)
      try {
        const response = await axios.post(
          'http://localhost:8080/validateLoads',
          data
        )
        if (response.status === 200) {
          this.setState({
            output: response.data.toString(),
          })
        }
      } catch (error) {
        this.setState({
          output: 'Wrong format file.',
        })
      }
    }
  }

  render() {
    return (
      <div>
        <input type="file" name="myFile" onChange={this.uploadFile} />
        <br></br>
        <br></br>
        <textarea value={this.state.output} style={textareaStyles} />
      </div>
    )
  }
}
