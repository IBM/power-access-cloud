import React, { useContext } from 'react';

const About = () => {
    const data = useContext((state)=>state);
    console.log(data);
  return (
    <div>About</div>
  )
}

export default About;
