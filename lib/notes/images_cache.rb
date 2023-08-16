class Notes::ImagesCache
  def get(url)
    cached_images.fetch(url) { process(url) }
  end

  private

  def cached_images
    File.exist?(index_path) ? read_index : {}
  end

  def read_index
    JSON.parse(File.read(index_path))
  rescue StandardError
    {}
  end

  def index_path
    @index_path ||= File.join(images_path, "index.json")
  end

  def process(url)
    image = Notes::CleanshotDownloader.new(url).download
    path = save_image(image)
    add_to_index(url, path)
    path
  end

  def add_to_index(url, path)
    File.write(index_path, JSON.pretty_generate(read_index.merge(url => path)))
  end

  def save_image(image)
    FileUtils.mkdir_p(images_path)
    file_name = new_file_name(File.extname(image[:file_name]))
    File.open(file_name, "wb") { image[:content] }
    File.basename(file_name)
  end

  def new_file_name(extension)
    File.join(images_path, "#{SecureRandom.uuid.gsub("-", "")}#{normalized_extension(extension)}")
  end

  def normalized_extension(extension)
    extension.downcase.then { _1 == ".jpeg" ? ".jpg" : _1 }
  end

  def images_path
    Notes::Configuration.images_path
  end
end
