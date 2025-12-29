import logo from "/logo.svg";

const Logo = () => {
  return (
    <img
      src={logo}
      className="h-8 hover:drop-shadow-[0_0_5px_rgba(60,130,240,1)] transition-all"
    />
  );
};

export default Logo;
