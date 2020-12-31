const expres= require("express");
const app= expres();
const port= process.env.port || 9000;
const bodyParser= require("body-parser");
const notifier = require("node-notifier");

app.use(bodyParser.json());


app.get("/health", (req,res)=> res.status(200).send());

app.post("/notify",(req,res)=>{
    notify(req.body, reply => res.send(reply));
})

app.listen(port,()=> console.log(`server is up on port ${port}`));

const notify = ({title, message},cb) =>{
    notifier.notify({
        title: title || "Unkown Title",
        message: message || "Unknown Message",
        sound:true,
        wait: true,
        reply: true,
        closeLabel: "Completed?",
        timeout: 15
    }, 
    (err,res,reply)=>{cb(reply)});
}