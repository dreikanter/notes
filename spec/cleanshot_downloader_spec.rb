RSpec.describe Notes::CleanshotDownloader do
  subject(:service_call) { described_class.new(cleanshot_url).download }

  let(:cleanshot_url) { "https://share.cleanshot.com/80085" }

  let(:direct_image_url) do
    "https://media.cleanshot.cloud/media/1334/itkYgfyMGKeaUT.jpeg?" \
    "response-content-disposition=attachment%3Bfilename%3DCleanShot" \
    "%25202023-08-06%2520at%252015.27.35.jpeg&Expires=1691367016"
  end

  let(:image_contents) { file_fixture("banana.jpg").read }

  context "with explicit file name" do
    subject(:service_call) { described_class.new(cleanshot_url, file_name: file_name).download }

    let(:file_name) { Tempfile.new(described_class.name) }

    before do
      stub_cleanshot_url.to_return(status: 302, headers: {"Location" => direct_image_url})
      stub_direct_image_url.to_return(headers: {"Content-Type" => "image/jpeg"}, body: image_contents)
      service_call
    end

    it { expect(File.read(file_name)).to eq(image_contents) }
  end

  context "with implicit file name" do
    before do
      stub_cleanshot_url.to_return(status: 302, headers: {"Location" => direct_image_url})
      stub_direct_image_url.to_return(headers: {"Content-Type" => "image/jpeg"}, body: image_contents)
    end

    it { expect(service_call).to eq("CleanShot 2023-08-06 at 15.27.35.jpeg") }
  end

  context "with redirect error" do
    before do
      stub_cleanshot_url.to_return(status: 500)
    end

    it { expect { service_call }.to raise_error("error getting image URL") }
  end

  context "with image downloading error" do
    before do
      stub_cleanshot_url.to_return(status: 302, headers: {"Location" => direct_image_url})
      stub_direct_image_url.to_return(status: 500)
    end

    it { expect { service_call }.to raise_error("error downloading image") }
  end

  def stub_cleanshot_url
    stub_request(:get, "#{cleanshot_url}+")
  end

  def stub_direct_image_url
    stub_request(:get, direct_image_url)
  end
end
